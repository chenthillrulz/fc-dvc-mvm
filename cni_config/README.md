## Installing CNI plugins
- Download cni plugin binaries and place them under /opt/cni/bin 
  - https://github.com/containernetworking/plugins/releases
  - Use version 1.0.0
- The tcp-redirect-tap cni plugin is a cni plugin from aws and is available at https://github.com/awslabs/tc-redirect-tap
  - Compile and place the binary at /opt/cni/bin 
- Place alpine.conflist under /etc/cni/conf.d/
   - alpine.conflist creates a bridge network
   - fcnet.conflist creates a p2p network
