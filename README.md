rulehuntersrv
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

Contributing
------------
If you want to improve this program make a pull request to the [repo](https://github.com/vlifesystems/rulehuntersrv) on github.  Please put any pull requests in a separate branch to ease integration and add a test to prove that it works.  If you find a bug, please report it at the project's [issues tracker](https://github.com/vlifesystems/rulehuntersrv/issues) also on github.

Installation
------------
Copy directories from `html_dependencies/` to the `wwwDir` directory specified in `config.json`.

Testing
-------
To test the output of the server you can create a simple static webserver using something like the following from the `wwwDir` directory specified in the `config.json`:

    ruby -run -ehttpd . -p8000


If you don't like ruby there is this [list of one-liner static webservers](https://gist.github.com/willurd/5720255).


Licence
-------
rulehuntersrv - A server to find rules in data based on user specified goals
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
This repository includes other packages in the `html_dependencies/` folder and in the `vendor/` folder.  Please see those packages for the licences that cover those works.
