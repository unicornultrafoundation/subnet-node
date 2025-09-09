package settings

const (
	BackupTargetSettingName       = "backup-target"
	VMForceResetPolicySettingName = "vm-force-reset-policy"

	AutoRotateRKE2CertsSettingName = "auto-rotate-rke2-certs"
	CSIDriverConfigSettingName     = "csi-driver-config"
)

var (
	provider Provider

	settings = map[string]Setting{}

	CSIDriverConfig = NewSetting(CSIDriverConfigSettingName, `{"driver.longhorn.io":{"volumeSnapshotClassName":"longhorn-snapshot","backupVolumeSnapshotClassName":"longhorn"}}`)
)

type Provider interface {
	Get(name string) string
	Set(name, value string) error
	SetIfUnset(name, value string) error
	SetAll(settings map[string]Setting) error
}

type Setting struct {
	Name     string
	Default  string
	ReadOnly bool
}

func NewSetting(name, def string) Setting {
	s := Setting{
		Name:    name,
		Default: def,
	}
	settings[s.Name] = s
	return s
}

func (s Setting) Get() string {
	if provider == nil {
		s := settings[s.Name]
		return s.Default
	}
	return provider.Get(s.Name)
}
