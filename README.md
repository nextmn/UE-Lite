# NextMN-UE Lite
**NextMN-UE Lite** is an experimental User Equipment simulator designed to be used along with **NextMN-gNB Lite** and **NextMN-CP Lite** to mimic from an UPF point-of-view a 5G & beyond Control Plane and a RAN.

3GPP N1/N2 interfaces are not (and will not be) implemented, and Control Plane is minimalistic on purpose.

This allow to test N3 and N4 interfaces of an UPF, and in particular to test handover procedures.

If you don't need to use handover procedures, consider using [UERANSIM](https://github.com/aligungr/UERANSIM) along with a real Control Plane (e.g. [free5GC](https://github.com/free5GC)'s NFs) instead.

## Getting started
### Build dependencies
- golang
- make (optional)

### Runtime dependencies
- iproute2
- iptables


### Build and install
Simply run `make build` and `make install`.

### Docker
If you plan using NextMN-UE Lite with Docker:
- The container required the `NET_ADMIN` capability;

This can be done in `docker-compose.yaml` by defining the following for the service:

```yaml
cap_add:
    - NET_ADMIN
```

## Author
Louis Royer

## Licence
MIT
