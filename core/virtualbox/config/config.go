package config

import (
	"path/filepath"

	"github.com/unicornultrafoundation/subnet-node/common/fsutil"
	"github.com/unicornultrafoundation/subnet-node/config"
)

type VBoxConfig struct {
	BaseFolder string `yaml:"base_folder"`
	ImagesDir  string `yaml:"images_dir"`
}

func New(cfg *config.C) *VBoxConfig {

	// ensure the base folder is a valid directory
	baseFolder, _ := fsutil.ExpandHome(cfg.GetString("virtualbox.base_folder", "~/VirtualBox VMs"))
	if err := fsutil.DirWritable(baseFolder); err != nil {
		return nil
	}

	imagesDir := filepath.Join(baseFolder, "Images")
	if err := fsutil.DirWritable(imagesDir); err != nil {
		return nil
	}

	return &VBoxConfig{
		BaseFolder: baseFolder,
		ImagesDir:  imagesDir,
	}

}
