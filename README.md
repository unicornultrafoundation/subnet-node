# Installing Subnet (Under Development)

This document provides instructions to install and run **Subnet**, whether you are building it from source or using prebuilt binaries. Note that this project is currently **under development**, and breaking changes may occur.

---

## **Table of Contents** 
<!-- no toc -->
  - [Using Prebuilt Binaries](#1-using-prebuilt-binaries)
  - [Building from Source](#2-building-from-source)
  - [Initializing Configuration](#3-initializing-configuration)
  - [Running the Subnet Node](#4-running-the-subnet-node)
  - [Development Status and Known Issues](#5-development-status-and-known-issues)
  - [Feedback and Contribution](#6-feedback-and-contribution)
  - [License](#license)
---

## **1. Using Prebuilt Binaries**

We provide prebuilt binaries for major platforms (Linux, macOS, and Windows). You can download the binary for your system and start using Subnet without building it from source.

### **Steps to Install and Run**

1. **Download the Prebuilt Binary**
   - Go to the [Releases Page](https://github.com/unicornultrafoundation/subnet-node/releases).
   - Download the binary for your operating system:
     - Linux: `subnet-linux-amd64`
     - macOS: `subnet-darwin-amd64`
     - Windows: `subnet-windows-amd64.exe`

2. **Make the Binary Executable (Linux/macOS)**
   ```bash
   chmod +x subnet-linux-amd64
   ```

3. **Move the Binary to Your PATH**
   ```bash
   sudo mv subnet-linux-amd64 /usr/local/bin/subnet
   ```

   For macOS, replace `subnet-linux-amd64` with `subnet-darwin-amd64`. For Windows, no additional steps are needed if you run the `.exe` file directly.

4. **Verify the Installation**
   ```bash
   subnet --version
   ```

5. **Initialize the Configuration**
   - Follow the [Initializing Configuration](#initializing-configuration) section.

---

## **2. Building from Source**

If you prefer to build Subnet from source, follow these steps:

### **Prerequisites**
- **Golang**: Version 1.19 or higher ([Install Go](https://golang.org/dl/)).
- **Git**: Version control system.

### **Steps to Build**

1. Clone the repository:
   ```bash
   git clone https://github.com/unicornultrafoundation/subnet-node.git
   cd subnet-node
   ```

2. Build the source code:
   ```bash
   go build -o ./build/subnet ./cmd/subnet
   ```

3. Verify the build:
   ```bash
   cd ./build
   subnet --help
   ```

---

## **3. Initializing Configuration**

Before running the Subnet node, you must initialize its configuration.

1. **Run the Initialization Command**
   ```bash
   subnet --init --datadir ~/.subnet
   ```

2. **Verify the Configuration**
   - Check the generated file at `~/.subnet/config.yaml`:
     ```yaml
     addresses:
         api:
         - /ip4/0.0.0.0/tcp/8080 # api server
         swarm:
         - /ip4/0.0.0.0/tcp/4001 # swarm server
      bootstrap:
         - /ip4/47.129.250.9/tcp/4001/p2p/12D3KooWDK63y6sxFi3dNqrS8yRetgbB81Tzszvs2yLoEtWtPCDa # subnet genesis boot node
      identity:
         peer_id: 12D3KooWLJCgSFv62DfQwNKTYxACMisLnCnGbYBp766p1K19PExx # DON'T USE! Example of ED25519 Public Key 
         privkey: CAESQA0P3Td0UZ2sAoCflMNdUivxFbUdcuo+XraGbZk5EZEdm7ZnWjxFiBUG9M718wpSOzl8P4JDQe6vZ6w+F5S9VPE= # DON'T USE! Example of ED25519 Private Key 
     ```

---

## **4. Running the Subnet Node**

Once the configuration is initialized, you can start the Subnet node.

1. **Start the Node**
   ```bash
   subnet --datadir ~/.subnet
   ```

2. **Connect to Bootstrap Peers**
   - If you need to connect to a bootstrap node manually:
     ```bash
     subnet connect /ip4/<bootstrap-ip>/tcp/<port>/p2p/<peer-id>
     ```

3. **Monitor Logs**
   - Observe logs for connections and resource discovery:
     ```plaintext
     Initializing Subnet Node...
     Node started with Peer ID: <ED25519 Public Key>
     Listening on: /ip4/127.0.0.1/tcp/4001
     ```

---

## **5. Development Status and Known Issues**

### **Current Status**
- **Active Development**: Subnet is currently in the alpha stage and under heavy development.
- **Breaking Changes**: APIs, configurations, and network protocols may change without notice.

### **Known Issues**
1. **Peer Connection Failures**:
   - NAT traversal issues may occur in certain network environments.
   - Temporary fix: Ensure ports are forwarded manually or run on public networks.

2. **Incomplete Features**:
   - Resource management and discovery are being actively developed.

3. **Stability**:
   - The system may become unstable under heavy load or high peer counts.

---

## **6. Feedback and Contribution**

We value your feedback to improve Subnet. If you encounter issues or have suggestions, please:

1. Open an issue on our [GitHub Repository](https://github.com/unicornultrafoundation/subnet-node/issues).
2. Join discussions and share ideas in the [Community Forum](https://github.com/unicornultrafoundation/subnet-node/discussions).

For contributing to Subnet, refer to [CONTRIBUTING.md](https://github.com/unicornultrafoundation/subnet-node/blob/main/CONTRIBUTING.md).

---

## **License**

This project is licensed under the [MIT License](./LICENSE).

---

For additional documentation, visit the [Wiki](https://github.com/unicornultrafoundation/subnet-node/wiki).
