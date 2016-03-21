[ ![Download][download-image] ][download-url]

Packer post-processor: VirtualBox to Hyper-V
---

A [packer] plugin to create Hyper-V vagrant boxes from VirtualBox artifacts.

You can use this plugin without having Hyper-V installed.


Usage
---

Add the plugin to your packer template:

```
{
  "builders": [
    {
      "type": "vmware-iso", ...
    },
    {
      "type": "virtualbox-iso", ...
    }
  ],
  "post-processors": [
    {
      "type": "vagrant",
      "output": "output_{{.Provider}}.box",
    },
    {
      "type": "virtualbox-to-hyperv",
      "only": ["virtualbox-iso"],
      "output": "output_hyperv.box"
    }
  ]
}
```

This would generate a vagrant box for vmware, virtualbox and hyperv providers

In order for a box to work in Hyper-V, the guest VM will need Hyper-V integration tools installed. Otherwise, vagrant will not be able to determine the guest's IP address.

For a real example, see this [packer template for Windows 2008 R2].

Configuration
---

This plugin extends the [Vagrant post-processor] and accepts the same configuration options.

It also accepts this additional configuration:
- `vm_name` (string) - the name of the virtual machine displayed in Hyper-V Manager.
  By default this is "packer-BUILDNAME", where "BUILDNAME" is the name of the build.
  Can be changed in the Vagrantfile by setting [vmname] in the hyperv provider.


Installation
---

1. [Download the plugin][download-url] into the same directory as your packer executable.
2. Add the plugin to `~/.packerconfig` or `%APPDATA%\packer.config` on Windows:

```
{
  "post-processors": {
    "virtualbox-to-hyperv": "packer-post-processor-virtualbox-to-hyperv"
  }
}
```

See also [Installing Plugins].


Credits
---

Based on [MSOpenTech/packer-hyperv] using the techniques described in [Creating a Hyper-V Vagrant box from a VirtualBox vmdk or vdi image]



[download-url]: https://bintray.com/dwickern/packer-plugins/packer-post-processor-virtualbox-to-hyperv/_latestVersion#files
[download-image]: https://api.bintray.com/packages/dwickern/packer-plugins/packer-post-processor-virtualbox-to-hyperv/images/download.svg
[Creating a Hyper-V Vagrant box from a VirtualBox vmdk or vdi image]: http://www.hurryupandwait.io/blog/creating-a-hyper-v-vagrant-box-from-a-virtualbox-vmdk-or-vdi-image
[packer]: https://www.packer.io/
[MSOpenTech/packer-hyperv]: https://github.com/MSOpenTech/packer-hyperv
[Installing Plugins]: https://www.packer.io/docs/extend/plugins.html
[Vagrant post-processor]: https://www.packer.io/docs/post-processors/vagrant.html
[vmname]: https://www.vagrantup.com/docs/hyperv/configuration.html
[packer template for Windows 2008 R2]: https://github.com/dwickern/packer-windows/blob/master/windows_2008_r2.json
