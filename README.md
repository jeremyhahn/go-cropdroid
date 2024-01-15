# CropDroid

Automated Indoor/Outdoor Farming

# Platform Support

|  | Standalone | Cluster |
| -- | :--: | :--: |
| **x86_64** | x | x |
| **ARM** | x | `-` |
| **ARM-64** | x | x |

# Dependencies

1. [packer-builder-arm-image](https://github.com/solo-io/packer-builder-arm-image)
2. [Screen](https://www.gnu.org/software/screen/)
3. [AVRDUDE](http://www.nongnu.org/avrdude/)
4. [Ansible] (https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html)

  # Required for cross compiling arm binary on x86_64
  sudo apt-get install gcc-multilib-arm-linux-gnueabihf

  # Build ARM image with packer
  sudo apt-get  kpartx qemu-user-static
  git clone https://github.com/solo-io/packer-builder-arm-image
  cd packer-builder-arm-image
  go mod download
  go build
  go build cmd/flasher/main.go

### Clustering Dependencies

1. [RocksDB](https://github.com/facebook/rocksdb/blob/master/INSTALL.md)


# Build

Use the provided Makefile to build the binary.

    # Build binaries for x64 and ARMv6
    make

### ARM

1. [CockroachDB](https://www.cockroachlabs.com/blog/run-cockroachdb-on-a-raspberry-pi/)
2. [Docker Buildx](https://www.docker.com/blog/getting-started-with-docker-for-arm-on-linux/)
3. [ARM64 QEMU](https://wiki.ubuntu.com/ARM64/QEMU)
4. [packer-builder-arm-image](https://github.com/solo-io/packer-builder-arm-image)
5. [qemu-rpi-kernel](https://github.com/dhruvvyas90/qemu-rpi-kernel)

### Cross Compiling

[Cross compiling](https://dh1tw.de/2019/12/cross-compiling-golang-cgo-projects/) for the Raspberry PI / ARM architecture on
Linux x64 requires the following toolchain:

  - gcc-6-arm-linux-gnueabihf

# Administration

1. [RocksDB Tooling](https://github.com/facebook/rocksdb/wiki/Administration-and-Data-Access-Tool)


# Databases

### Sqlite
uint64 not supported: https://github.com/golang/go/issues/9373


# Devices

## Room Device

### Networking

    # Default MAC and IP
    MAC: 0x04 0x02 0x00 0x00 0x01
    IP: 192.168.0.91

    # Set new MAC and IP
    CONTROLLER=192.168.0.91
    curl http://$CONTROLLER/eeprom/0/04
    curl http://$CONTROLLER/eeprom/1/02
    curl http://$CONTROLLER/eeprom/2/00
    curl http://$CONTROLLER/eeprom/3/00
    curl http://$CONTROLLER/eeprom/4/01
    curl http://$CONTROLLER/eeprom/5/01
    curl http://$CONTROLLER/eeprom/6/192
    curl http://$CONTROLLER/eeprom/7/168
    curl http://$CONTROLLER/eeprom/8/0
    curl http://$CONTROLLER/eeprom/9/51
    curl http://$CONTROLLER/reboot

### Switches

1. Lighting      (120 VAC)
2. Dehumidifier  (120 VAC)
3. Ventilation   (120 VAC)
4. A/C           (120 VAC)
5. Heater        (120 VAC)
6. CO2           (120 VAC)
7. Auxiliary     (120 VAC)
8. Auxiliary     (120 VAC)

### Sensors
1. Pod0 upper float
2. Pod0 water temp (DS18B20)
3. Pod1 upper float
4. Pod1 water temp (DS18B20)
5. Analog water sensors (floor)
6. DHT22 (x3)(ceiling, floor, canopy)
7. CO2
8. Photo (light)
9. Smoke

## Reservoir Device

### Networking

    # Default MAC and IP
    MAC: 0x04 0x02 0x00 0x00 0x02
    IP: 192.168.0.92

    # Set new MAC and IP
    CONTROLLER=192.168.0.92
    curl http://$CONTROLLER/eeprom/0/04
    curl http://$CONTROLLER/eeprom/1/02
    curl http://$CONTROLLER/eeprom/2/00
    curl http://$CONTROLLER/eeprom/3/00
    curl http://$CONTROLLER/eeprom/4/01
    curl http://$CONTROLLER/eeprom/5/01
    curl http://$CONTROLLER/eeprom/6/192
    curl http://$CONTROLLER/eeprom/7/168
    curl http://$CONTROLLER/eeprom/8/0
    curl http://$CONTROLLER/eeprom/9/52
    curl http://$CONTROLLER/reboot

### Switches

1. Pod0 pump   (120 VAC)
2. Pod1 pump   (120 VAC)
3. Heater      (120 VAC)
4. Chiller     (120 VAC)
5. Drain       (12 VDC)
6. Fill        (12 VDC)
7. Top-Off     (12 VDC)

### Sensors

1. Upper float
2. Lower float
3. Water Temp (DS18B20)
4. DHT22
5. PH
6. EC
7. DO
8. ORP

## Dosing Device

### Networking

    # Default MAC and IP
    MAC: 0x04 0x02 0x00 0x00 0x03
    IP: 192.168.0.93

    # Set new MAC and IP
    CONTROLLER=192.168.0.93
    curl http://$CONTROLLER/eeprom/0/04
    curl http://$CONTROLLER/eeprom/1/02
    curl http://$CONTROLLER/eeprom/2/00
    curl http://$CONTROLLER/eeprom/3/00
    curl http://$CONTROLLER/eeprom/4/01
    curl http://$CONTROLLER/eeprom/5/01
    curl http://$CONTROLLER/eeprom/6/192
    curl http://$CONTROLLER/eeprom/7/168
    curl http://$CONTROLLER/eeprom/8/0
    curl http://$CONTROLLER/eeprom/9/53
    curl http://$CONTROLLER/reboot

## CropDroid Main Device


## Raft

```
As a summary, when -
  - starting a brand new Raft cluster, set join to false and specify all initial
    member node details in the initialMembers map.
  - joining a new node to an existing Raft cluster, set join to true and leave
    the initialMembers map empty. This requires the joining node to have already
    been added as a member node of the Raft cluster.
  - restarting an crashed or stopped node, set join to false and leave the
    initialMembers map to be empty. This applies to both initial member nodes
    and those joined later.
```

## Safety Requirements and Standards

- IPC-2221 Voltage and Spacing Standards
- IEC-60950-1 (2nd edition)


## RPI Touchscreen
https://www.amazon.com/gp/product/B07NRYPZM1
http://www.lcdwiki.com/5inch_HDMI_Display 


# Stores

## Config store

The config store is responsible for storing desired device configurations and/or settings.

## State Store

The state store is responsible for keeping a single "record" for a device that indicates it's current metric and channel values.

## Device Store

The device store is responsible for keeping historical data for a device - that is, many metric and channel values over time, used for analytics and reporting.


# Golang channels

The following golang channels are created for each farm

1. Farm state gc ticker
2. Device state gc ticker
3. farm.WatchFarmStateChange
4. farm.WatchFarmConfigChange
5. farm.WatchDeviceStateChange


# Known Bugs

- Not getting push notifications
- Enable/disable toggles wrong sensor
- Conditions not ordered by time
- Workflow steps out of order
- Roles not loading
