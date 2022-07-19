# rubik

## Description

The current global cloud infrastructure service expenditure is huge. However, the average CPU utilization of data center user clusters is very low, which is a huge waste of resources. Therefore, improving the utilization of data center resources is an important issue that needs to be solved urgently.

Deployment of various types of services togather can significantly improve the utilization of cluster resources, but it also brings the problem of co-peaking, this issue can lead to partial service quality of service (QoS) compromise. How to ensure that the application's QoS is not damaged after improving the utilization of resources is a key technical challenge.

To this end, we propose the Rubik resource utilization improvement solution, Rubik literally means Rubik's Cube, The Rubik’s Cube was invented in 1974 by Ernõ Rubik, a Hungarian architecture professor. In our solution, Rubik symbolizes being able to manage servers in an orderly manner.

Rubik currently supports the following features:
- pod's CPU priority configure.
- pod's memory priority configure.

## Build

Pull the source code:
```sh
git clone https://gitee.com/openeuler/rubik.git
```

Enter the source code directory to compile:
```sh
cd rubik
make
```

Make rubik image:
```sh
make image
```

Install the relevant deployment files on the system:
```sh
sudo make install
```
## Deployment

### Prepare environment

- OS: openeuler 21.09/22.03/22.09+
- kubernetes: 1.17.0+

### Deploy rubik as daemonset

```sh
kubectl apply -f /var/lib/rubik/rubik-daemonset.yaml
```

## Copyright

Mulan PSL v2
