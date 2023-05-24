1. Architecture optimization:
refactor rubik through `informer-podmanager-services` mechanism, decoupling modules and improving performance
2. Interface change:
- configuration file changes
- use the list-watch mechanism to get the pod instead of the http interface
3. Feature enhancements:
- support elastic cpu limit user mode scheme-quotaturbo
- support psi index observation
- support memory asynchronous recovery feature (fssr optimization)
- support memory access bandwidth and LLC limit
- optimize the absolute preemption
- optimize the elastic cpu limiting kernel mode scheme-quotaburst
4. Other optimizations:
- document optimization
- typo fix
- compile option optimization
