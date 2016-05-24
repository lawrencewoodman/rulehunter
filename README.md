Rulehuntersrv
=============
A server to find rules in data based on user specified goals

Requirements
------------
These have all be vendored into the `vendor/` directory using the
[govendor](https://github.com/kardianos/govendor) tool.

* [dexpr](https://github.com/lawrencewoodman/dexpr) package
* [dlit](https://github.com/lawrencewoodman/dlit) package
* [rulehunter](https://github.com/vlifesystems/rulehunter) package
* [osext](https://github.com/kardianos/osext) package
* [service](https://github.com/kardianos/service) package


To format the html pages properly [Twitter Boostrap](http://getbootstrap.com) is used and hence has been included in the `support/` directory.

Contributing
------------
[![Build Status](https://travis-ci.org/vlifesystems/rulehuntersrv.svg?branch=master)](https://travis-ci.org/vlifesystems/rulehuntersrv)

If you want to improve this program make a pull request to the [repo](https://github.com/vlifesystems/rulehuntersrv) on github.  Please put any pull requests in a separate branch to ease integration and add a test to prove that it works.  If you find a bug, please report it at the project's [issues tracker](https://github.com/vlifesystems/rulehuntersrv/issues) also on github.

Installation
------------
Copy directories from `support/bootstrap` and `support/rulehuntersrv` to the `wwwDir` directory specified in `config.json`.

Testing
-------
To test the output of the server you can create a simple static webserver using something like the following from the `wwwDir` directory specified in the `config.json`:

    ruby -run -ehttpd . -p8000


If you don't like ruby there is this [list of one-liner static webservers](https://gist.github.com/willurd/5720255).


Licence
-------
Rulehuntersrv - A server to find rules in data based on user specified goals

Copyright (C) 2016 [vLife Systems Ltd](http://vlifesystems.com)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program; see the file COPYING.  If not, see
[http://www.gnu.org/licenses/](http://www.gnu.org/licenses/).

Additional Licences
-------------------

### Twitter Boostrap

This has been included in the repository under `support/bootstrap`.

Copyright 2011-2015 Twitter, Inc.

Licensed under the MIT license.  For details see: [http://getbootstrap.com](http://getbootstrap.com).

### Vendored Go Packages

This repository includes other packages in the `vendor/` directory.  Please see those packages for the licences that cover those works.
