package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"packer-post-processor-virtualbox-to-hyperv/hyperv"

	vbox "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/vagrant"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	vagrant.Config `mapstructure:",squash"`

	StagingDir string `mapstructure:"staging_directory"`
	VMName     string `mapstructure:"vm_name"`
	DiskName   string `mapstructure:"disk_name"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config

	// embed the "vagrant" post-processor to work around https://github.com/mitchellh/packer/issues/1653
	vagrant vagrant.PostProcessor

	// used to convert the hard drive format
	virtualbox vbox.Driver
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// configure defaults
	if p.config.StagingDir == "" {
		p.config.StagingDir = fmt.Sprintf("output-%s-hyperv", p.config.PackerBuildName)
	}
	if p.config.VMName == "" {
		p.config.VMName = fmt.Sprintf("packer-%s-%d", p.config.PackerBuildName, interpolate.InitTime.Unix())
	}
	if p.config.DiskName == "" {
		p.config.DiskName = fmt.Sprintf("packer-%s-%d", p.config.PackerBuildName, interpolate.InitTime.Unix())
	}

	// configure "vagrant" post-processor
	if err := p.vagrant.Configure(raws...); err != nil {
		return fmt.Errorf("Failed to configure vagrant post-processor: %s", err)
	}

	// find the VBoxManage executable
	if p.virtualbox, err = vbox.NewDriver(); err != nil {
		return fmt.Errorf("Failed to create VirtualBox driver: %s", err)
	}

	if _, err := os.Stat(p.config.StagingDir); err == nil {
		return fmt.Errorf("Staging directory '%s' already exists.", p.config.StagingDir)
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != "mitchellh.virtualbox" {
		err := fmt.Errorf("Unknown artifact type: %s\nCan only import from virtualbox artifacts.", artifact.BuilderId())
		return nil, false, err
	}

	outVirtualMachines := filepath.Join(p.config.StagingDir, "Virtual Machines")
	if err := os.MkdirAll(outVirtualMachines, 0755); err != nil {
		return nil, false, fmt.Errorf("Failed to create output directory: %s", err)
	}

	outVirtualHardDisks := filepath.Join(p.config.StagingDir, "Virtual Hard Disks")
	if err := os.MkdirAll(outVirtualHardDisks, 0755); err != nil {
		return nil, false, fmt.Errorf("Failed to create output directory: %s", err)
	}

	vmdk, err := FindVirtualHardDisk(artifact)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to find VirtualBox hard disk: %s", err)
	}

	vhd, err := filepath.Abs(filepath.Join(outVirtualHardDisks, fmt.Sprintf("%s.vhd", p.config.DiskName)))
	if err != nil {
		return nil, false, fmt.Errorf("Failed to get absolute path to VHD: %s", err)
	}

	if err = p.virtualbox.VBoxManage("clonehd", vmdk, vhd, "--format", "vhd"); err != nil {
		return nil, false, fmt.Errorf("Failed to convert VMDK to VHD: %s", err)
	}

	vm := filepath.Join(outVirtualMachines, "vm.xml")
	if err := ioutil.WriteFile(vm, p.CreateVM(vhd), 0644); err != nil {
		return nil, false, fmt.Errorf("Failed to write vm.xml: %s", err)
	}

	newArtifact, err := hyperv.NewArtifact(p.config.StagingDir)
	if err != nil {
		return nil, false, fmt.Errorf("Failed to create Hyper-V artifact: %s", err)
	}

	return p.vagrant.PostProcessProvider("hyperv", new(hyperv.HypervProvider), ui, newArtifact)
}

func (p *PostProcessor) CreateVM(vhd string) []byte {
	vm := VM_XML
	vm = strings.Replace(vm, "INSERT_VM_NAME_HERE", p.config.VMName, 1)
	vm = strings.Replace(vm, "INSERT_VHD_PATH_HERE", vhd, 1)
	// TODO: update `creation_time`
	return []byte(vm)
}

func FindVirtualHardDisk(artifact packer.Artifact) (string, error) {
	for _, path := range artifact.Files() {
		if strings.ToLower(filepath.Ext(path)) != ".vmdk" {
			continue
		}

		abs, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("Failed to get absolute path to VMDK: %s", err)
		}

		return abs, nil
	}

	return "", errors.New("Input artifact did not have a VMDK file")
}
