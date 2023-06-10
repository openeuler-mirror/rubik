# CHANGELOG

## v2.0.1

### New Feature

Before June 30, 2023

1. **dynMemory** (asynchronous memory classification recovery): implement fssr strategy
2. **psi**: interference detection based on PSI index
3. **quotaTurbo**: elastic cpu limit user mode solution

## v2.0.0

### Architecture optimization

refactor rubik through `informer-podmanager-services` mechanism, decoupling modules and improving performance

### Interface change

- configuration file changes
- use the list-watch mechanism to get the pod instead of the http interface

### Feature enhancements

- support elastic cpu limit user mode scheme-quotaturbo
- support psi index observation
- support memory asynchronous recovery feature (fssr optimization)
- support memory access bandwidth and LLC limit
- optimize the absolute preemption
- optimize the elastic cpu limiting kernel mode scheme-quotaburst

### Other optimizations

- document optimization
- typo fix
- compile option optimization
