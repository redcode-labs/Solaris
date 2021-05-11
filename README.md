
<p align="center">
  <a>
    <img src="solaris.png" width="700">
  </a>
</p>

<p align="center">
  A local LKM rootkit loader
</p>


![Language](https://img.shields.io/badge/Language-Go-blue.svg?longCache=true&style=flat-square)   ![License](https://img.shields.io/badge/License-MIT-purple.svg?longCache=true&style=flat-square)  


## Introduction
This loader can list both user and kernel mode protections that are present on the system, and additionally disable some of them.

It locally drops and compiles source code of any Linux kernel-mode rootkit specified by the user.

## Usage

<p align="center">
  <a>
    <img src="usage.png" width="790">
  </a>
</p>

Place the code of your selected rootkit inside `rootkit_template` variable within `solaris.go`. 

Compile the Golang binary and launch it on the target system.

## License

This software is under [MIT License](https://en.wikipedia.org/wiki/MIT_License)