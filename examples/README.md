Examples
========

This directory (`examples/`) contains example experiments:

    experiments/acmeprinters_repair016_who_should_call_which_segment.yaml
    experiments/breast_cancer_wisconsin_benign_high.yaml
    experiments/breast_cancer_wisconsin_malignant.yaml
    experiments/breast_cancer_wisconsin_malignant_filtered.yaml
    experiments/breast_cancer_wisconsin_malignant_high.yaml
    experiments/iris-setosa.yaml
    experiments/iris-versicolor.yaml
    experiments/iris-virginica.yaml

Preparing the `www` Directory
-----------------------------
Before running `rulehunter` on the experiments you need to initialize the
`www` directory.

For Linux/MacOS:

```Shell
    chmod +x examples/bin/init_www_unix.sh
    examples/bin/init_www_unix.sh
```

For Windows:

```Shell
    examples\bin\init_www_windows.bat
```

Processing the Experiments
--------------------------

To process the experiments run `rulehunter` from the `examples/` directory:

```Shell
    rulehunter
```

The website will be generated in the `examples/www` directory and can
be viewed with a simple static webserver such as the following run from
the `examples/www` directory:

```Shell
ruby -run -ehttpd . -p8000
```

If you don't like ruby there is this [list of one-liner static webservers](https://gist.github.com/willurd/5720255).
