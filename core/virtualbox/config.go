package virtualbox

import (
	"path/filepath"

	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/config"
)

type VirtualBoxConfig struct {
	BaseFolder string `yaml:"base_folder"`
	ImagesDir  string `yaml:"images_dir"`
}

func GetVirtualBoxConfig(cfg *config.C) (VirtualBoxConfig, error) {

	// ensure the base folder is a valid directory
	baseFolder, err := fsutil.ExpandHome(cfg.GetString("virtualbox.base_folder", "~/VirtualBox VMs"))
	if err != nil {
		return VirtualBoxConfig{}, err
	}
	if err := fsutil.DirWritable(baseFolder); err != nil {
		return VirtualBoxConfig{}, err
	}

	imagesDir := filepath.Join(baseFolder, "Images")
	if err := fsutil.DirWritable(imagesDir); err != nil {
		return VirtualBoxConfig{}, err
	}

	return VirtualBoxConfig{
		BaseFolder: baseFolder,
		ImagesDir:  imagesDir,
	}, nil
}
