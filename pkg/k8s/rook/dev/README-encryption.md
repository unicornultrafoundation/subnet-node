# Rook Ceph CSI Encryption Setup

This directory contains the configuration files for enabling encryption on Rook Ceph CSI volumes.

## Files Updated

### 1. Operator Configuration
- `operator.yaml` - Already has `CSI_ENABLE_ENCRYPTION: "true"`

### 2. StorageClasses
- `csi/rbd/storageclass.yaml` - Updated with `encryptionKMSID: "user-secret-metadata"`
- `csi/cephfs/storageclass.yaml` - Updated with `encryptionKMSID: "user-secret-metadata"`

### 3. KMS Configuration
- `kms-config.yaml` - Contains the KMS ConfigMap and Secret for metadata-based encryption

### 4. RBAC Configuration
- `../core/csi-encryption-rbac.yaml` - Required ClusterRole and ClusterRoleBinding for CSI components to read encryption secrets

## Encryption Configuration

The encryption uses a **metadata-based approach** as documented in the official Rook documentation:

```yaml
# KMS ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: rook-ceph-csi-kms-config
  namespace: rook-ceph
data:
  config.json: |-
    {
      "user-secret-metadata": {
        "encryptionKMSType": "metadata",
        "secretName": "storage-encryption-secret"
      }
    }

# Encryption Secret
apiVersion: v1
kind: Secret
metadata:
  name: storage-encryption-secret
  namespace: default  # Must be in the same namespace as PVCs
stringData:
  encryptionPassphrase: "ChangeMe-Strong-Secret-123!"
```

## StorageClass Configuration

Both RBD and CephFS StorageClasses are configured with:

```yaml
parameters:
  encrypted: "true"
  encryptionKMSID: "user-secret-metadata"
```

## Deployment Steps

1. **Apply the KMS configuration:**
   ```bash
   kubectl apply -f kms-config.yaml
   ```

2. **Deploy Rook with encryption enabled:**
   ```bash
   kubectl apply -f operator.yaml
   kubectl apply -f cluster.yaml
   kubectl apply -f csi/
   ```

3. **Create encrypted PVCs:**
   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: encrypted-pvc
   spec:
     accessModes: ["ReadWriteOnce"]
     storageClassName: rook-ceph-block  # or rook-cephfs
     resources:
       requests:
         storage: 1Gi
   ```

## Important Notes

- **Secret Location**: The `storage-encryption-secret` must be in the **same namespace** as the PVCs
- **Encryption Type**: Uses `encryptionKMSType: "metadata"` (not `kms-kubernetes`)
- **LUKS Encryption**: RBD volumes are encrypted using LUKS
- **fscrypt**: CephFS volumes use fscrypt (requires kernel 6.6+)
- **RBAC Required**: CSI components need ClusterRole permissions to read encryption secrets across namespaces

## Verification

To verify encryption is working:

1. Check PVC status: `kubectl get pvc <name>`
2. Verify volume attributes: `kubectl get pv <pv-name> -o yaml`
3. Check RBD image: `rbd info <pool>/<image-name>`
4. Mount in pod: Should show `/dev/mapper/luks-rbd-*` for RBD volumes

## Security Considerations

- Change the default passphrase in production
- Store the passphrase securely
- Consider using a proper KMS (Vault, AWS KMS, etc.) for production
- The current setup uses Kubernetes secrets for simplicity
