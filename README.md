# Firecracker Demo
- Read the cni_config README to setup cni plugins and network for running firecracker demo
  - https://github.com/chenthillrulz/fc-dvc-mvm/blob/master/cni_config/README.md 
- Edit kernelImagePath, rootFSPath, connectorFSPath to set appropriate paths
- Set the path for firecracker executable file in PATH environment variable
- Compile and run the Go binary to create the microVM
- The Nodejs app runs at http://\<vm-ip-address\>:3000
