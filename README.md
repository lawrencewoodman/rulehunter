Rulehunter
==========

[![Build Status](https://travis-ci.org/vlifesystems/rulehunter.svg?branch=master)](https://travis-ci.org/vlifesystems/rulehunter)
[![Build status](https://ci.appveyor.com/api/projects/status/8tds5r4dk6163es0?svg=true)](https://ci.appveyor.com/project/lawrencewoodman/rulehunter)
[![Coverage Status](https://coveralls.io/repos/vlifesystems/rulehunter/badge.svg?branch=master)](https://coveralls.io/r/vlifesystems/rulehunter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/vlifesystems/rulehunter)](https://goreportcard.com/report/github.com/vlifesystems/rulehunter)

Rulehunter allows you to describe your goals and will then find simple rules to use on a supplied dataset to meet those goals.

Getting Started
---------------
* Follow the Rulehunter [installation](http://rulehunter.com/docs/installation/) instructions.  There are prebuilt binaries for a number of platforms.
* Read about its [configuration](http://rulehunter.com/docs/configuration/) and [usage](http://rulehunter.com/docs/usage/)
* Start creating [experiments](http://rulehunter.com/docs/experiments/)
* See the [Rulehunter](http://rulehunter.com) website for more information

Testing
-------
To make testing simpler under Linux, where root is needed, you can use the following (replace _systemd_ with _upstart_ if using the latter init system):

```Shell
sudo ./linux-test-su.sh $GOPATH `which go` systemd
```

To test the output of the server you can create a simple static webserver using something like the following from the `wwwDir` directory specified in the `config.yaml`:

```Shell
ruby -run -ehttpd . -p8000
```

If you don't like ruby there is this [list of one-liner static webservers](https://gist.github.com/willurd/5720255).

Contributing
------------
We would love contributions to improve this project.  You can help by reporting bugs, improving the documentation, submitting feature requests, fixing bugs, etc.

Please see the [Contributing Guide](https://github.com/vlifesystems/rulehunter/blob/master/CONTRIBUTING.md) for more details.

Reporting Issues
----------------
We want your help to find bugs as quickly as possible and would welcome your help by reporting any found.  Please see the [Contributing Guide](https://github.com/vlifesystems/rulehunter/blob/master/CONTRIBUTING.md) for more details of how to report them.

Licence
-------

Rulehunter - Find rules in data based on user specified goals

Copyright (C) 2016-2017 [vLife Systems Ltd](http://vlifesystems.com)

Rulehunter is licensed under the GNU Affero General Public License version 3 (AGPLv3). Please see the [licence](https://github.com/vlifesystems/rulehunter/blob/master/CONTRIBUTING.md#licence) section of the [Contributing Guide](https://github.com/vlifesystems/rulehunter/blob/master/CONTRIBUTING.md) for more details about this and the licences used by other code included with Rulehunter.
