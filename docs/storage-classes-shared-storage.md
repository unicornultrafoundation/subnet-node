# Storage Classes and Shared Storage Implementation

## 📚 Overview

This document explains how storage classes work in the subnet-node system and how to implement shared storage for applications like MinIO that need multiple pods to access the same data.

## 🔍 Storage Classes Explained

### What is a Storage Class?

A **Storage Class** in Kubernetes defines:
- **Provisioner**: How the storage is created (e.g., `rancher.io/local-path`)
- **Access Modes**: How pods can access the storage
- **Reclaim Policy**: What happens when the PVC is deleted
- **Volume Binding Mode**: When the volume is created

### Access Modes

| Access Mode | Description | Use Case |
|-------------|-------------|----------|
| **ReadWriteOnce (RWO)** | Only one pod can mount the volume | Single instance applications |
| **ReadWriteMany (RWM)** | Multiple pods can mount the volume | Shared storage applications |
| **ReadOnlyMany (ROM)** | Multiple pods can read, none can write | Shared read-only data |

### Current Storage Classes

```yaml
# Available storage classes in the system
- name: default
  provisioner: rancher.io/local-path
  accessModes: [ReadWriteOnce]
  
- name: local-path
  provisioner: rancher.io/local-path  
  accessModes: [ReadWriteOnce]
  
- name: v1
  provisioner: rancher.io/local-path
  accessModes: [ReadWriteOnce]
  
- name: shared-storage  # NEW!
  provisioner: rancher.io/local-path
  accessModes: [ReadWriteMany]  # Supports shared access
```

## 🔧 Implementation Details

### 1. Storage Class Configuration

**File**: `pkg/k8s/apis/storageclass.yaml`

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: shared-storage
  labels:
    subnet.node: "true"
    app.kubernetes.io/name: "shared-storage"
provisioner: rancher.io/local-path
reclaimPolicy: Retain
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
```

### 2. PVC Creation Logic

**File**: `core/k8s/kube/builder/workload.go`

The system automatically determines access modes based on storage class:

```go
// Determine access mode based on storage class
accessModes := []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}

// Check if this is a shared storage class that supports ReadWriteMany
sharedStorageClasses := map[string]bool{
    "shared-storage": true,
    "nfs":            true,
    "ceph":           true,
    "glusterfs":      true,
}

if sharedStorageClasses[class] {
    accessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany}
}
```

### 3. Allowed Storage Classes

**File**: `pkg/k8s/sdl/storage.go`

```go
var allowedStorageClasses = map[string]bool{
    "default":        true,
    "beta1":          true,
    "beta2":          true,
    "beta3":          true,
    "shared-storage": true,  // NEW!
    "nfs":            true,  // NEW!
    "ceph":           true,  // NEW!
    "glusterfs":      true,  // NEW!
    StorageClassRAM:  true,
}
```

## 🎯 MinIO Shared Storage Example

### Problem with Current Setup

Your current MinIO deployment has:
- **2 PVCs**: Separate volumes for each pod
- **ReadWriteOnce**: Only one pod can mount each volume
- **No data sharing**: Each pod has independent data

### Solution: Shared Storage

**File**: `core/deployer/manifest/examples/minio-shared-storage.json`

```json
{
  "version": "2.0",
  "services": {
    "minio": {
      "image": "minio/minio:latest",
      "command": ["minio", "server", "/data"],
      "params": {
        "storage": {
          "shared-data": {
            "mount": "/data"
          }
        }
      },
      "count": 2
    }
  },
  "profiles": {
    "compute": {
      "shared": {
        "resources": {
          "storage": [
            {
              "name": "shared-data",
              "size": "10Gi",
              "attributes": {
                "persistent": true,
                "class": "shared-storage"  // This enables ReadWriteMany
              }
            }
          ]
        }
      }
    }
  },
  "deployment": {
    "minio": {
      "westcoast": {
        "profile": "shared",
        "count": 2
      }
    }
  }
}
```

## ✅ Benefits of Shared Storage

### For MinIO with Shared Storage:

1. **✅ Session Persistence**: All pods share the same session data
2. **✅ Data Consistency**: Files uploaded to any pod are visible to all pods
3. **✅ Configuration Sharing**: Buckets and settings are synchronized
4. **✅ High Availability**: Multiple pods can serve requests
5. **✅ Load Balancing**: Traffic can be distributed across pods

### How It Works:

1. **Single PVC**: One `PersistentVolumeClaim` with `ReadWriteMany` access
2. **Multiple Mounts**: All pods mount the same volume
3. **Shared Data**: All data is stored in the same location
4. **Synchronized State**: MinIO instances share configuration and sessions

## 🚀 Usage Examples

### Example 1: Basic Shared Storage

```json
{
  "profiles": {
    "compute": {
      "shared": {
        "resources": {
          "storage": [
            {
              "name": "shared-data",
              "size": "10Gi",
              "attributes": {
                "persistent": true,
                "class": "shared-storage"
              }
            }
          ]
        }
      }
    }
  }
}
```

### Example 2: Multiple Shared Volumes

```json
{
  "profiles": {
    "compute": {
      "multi-shared": {
        "resources": {
          "storage": [
            {
              "name": "data",
              "size": "10Gi",
              "attributes": {
                "persistent": true,
                "class": "shared-storage"
              }
            },
            {
              "name": "cache",
              "size": "5Gi",
              "attributes": {
                "persistent": true,
                "class": "shared-storage"
              }
            }
          ]
        }
      }
    }
  }
}
```

### Example 3: Service Configuration

```json
{
  "services": {
    "app": {
      "params": {
        "storage": {
          "data": {
            "mount": "/app/data"
          },
          "cache": {
            "mount": "/app/cache"
          }
        }
      }
    }
  }
}
```

## ⚠️ Important Notes

### 1. Provisioner Limitations

The `rancher.io/local-path` provisioner **does not support ReadWriteMany** in production. For true shared storage, you need:

- **NFS**: Network File System
- **Ceph**: Distributed storage system
- **GlusterFS**: Distributed file system
- **Cloud Storage**: AWS EFS, Azure Files, GCP Filestore

### 2. Performance Considerations

- **Network Storage**: Shared storage may have higher latency
- **Concurrent Access**: Multiple pods writing simultaneously may cause conflicts
- **File Locking**: Applications need to handle concurrent file access

### 3. Backup and Recovery

- **Single Point of Failure**: All pods depend on the same storage
- **Backup Strategy**: Need to backup the shared volume
- **Disaster Recovery**: Plan for storage failure scenarios

## 🔄 Migration Guide

### From Separate Storage to Shared Storage

1. **Backup Data**: Export data from current MinIO instances
2. **Update Manifest**: Change storage class to `shared-storage`
3. **Redeploy**: Apply the new configuration
4. **Restore Data**: Import data to the shared storage
5. **Verify**: Test session persistence and data sharing

### Example Migration

**Before (Separate Storage)**:
```json
{
  "attributes": {
    "persistent": true,
    "class": "default"  // ReadWriteOnce
  }
}
```

**After (Shared Storage)**:
```json
{
  "attributes": {
    "persistent": true,
    "class": "shared-storage"  // ReadWriteMany
  }
}
```

## 🧪 Testing Shared Storage

### 1. Deploy Test Application

```bash
# Deploy MinIO with shared storage
kubectl apply -f minio-shared-storage.json
```

### 2. Verify PVC Creation

```bash
# Check PVC access modes
kubectl get pvc -o wide
```

### 3. Test Data Sharing

```bash
# Create file in pod 1
kubectl exec -it pod/minio-0 -- touch /data/test.txt

# Verify file exists in pod 2
kubectl exec -it pod/minio-1 -- ls -la /data/test.txt
```

### 4. Test Session Persistence

1. Access MinIO console from pod 1
2. Create a bucket
3. Access MinIO console from pod 2
4. Verify the bucket exists

## 📋 Troubleshooting

### Common Issues

1. **PVC Pending**: Storage class not available
2. **Mount Failures**: Provisioner doesn't support ReadWriteMany
3. **Permission Errors**: File system permissions
4. **Performance Issues**: Network storage latency

### Debug Commands

```bash
# Check storage classes
kubectl get storageclass

# Check PVC status
kubectl get pvc

# Check pod events
kubectl describe pod <pod-name>

# Check volume mounts
kubectl exec -it <pod-name> -- df -h
```

## 🎯 Conclusion

Shared storage enables:
- **Session persistence** across multiple pods
- **Data consistency** in distributed applications
- **High availability** with load balancing
- **Simplified backup** with single storage location

For MinIO specifically, shared storage solves the session persistence issue and allows multiple replicas to work together effectively.



