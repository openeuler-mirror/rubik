## How to Construct the go fuzz test

> Note
> Before you start, make sure you have `go-fuzz-build` and `go-fuzz` binaries, you can get them at [go-fuzz](https://github.com/dvyukov/go-fuzz)

1. cd `rubik/tests/data/fuzz-test` and make folder the form like `fuzz-test-xxx`
2. put the materials used by fuzz in the folder you created, they looks like the following form:
```bash
$ tree fuzz-test-newconfig
   fuzz-test-newconfig		# test case root dir
   |-- corpus				# dir to store mutation corpus
   |   |-- case1            # mutation corpus1
   |   |-- case2            # mutation corpus2
   |   |-- case3            # mutation corpus3
   |-- Fuzz                 # fuzz go file
   |-- path                 # record relative path to put the Fuzz file in the package
```
3. when the above meterials are ready, go to `rubik/tests/src`
4. the **ONLY Three Things** you need to do is:
    1. copy `TEMPLATE` file to the name you want(*must start with `fuzz_test`*), for example `fuzz_test_xxx.sh`
    2. change the variable `test_name` in the script you just copy same as the name you just gave(keep same with the folder you create in the first step)
    3. uncomment the last line `main "$1"`
5. To run single go fuzz shell script by doing `$ bash fuzz_test_xxx.sh`, it will stop fuzzing after 1 minute.
   If you want to change the default run time, you could do like `$ bash fuzz_test_xxx.sh 2h` to keep running 2 hours
6. To run **all** go fuzz shell scripts by first go to `rubik/tests`, then run `$ bash test.sh 2h`.
   It will run all go fuzz testcases and will stop fuzzing after `2h * number of go fuzz testcases`
