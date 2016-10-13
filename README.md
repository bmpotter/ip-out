# ip-out

## Introduction

This project comprises a small, standalone binary for discoverying and outputting network interfaces on a Linux system. It is intended for consumption by users of (primarily) headless Horizon system images during setup; it is executed upon boot on such images and its output is displayed on system setup web pages.

Related Projects:

* `ubuntu-classic-image` (http://github.com/open-horizon/ubuntu-classic-image): Produces complete system images

## Operations

#### Execution

Pretty-print discovered network interfaces:

    ip-out -pp

Output discovered interfaces in JSON format:

    ip-out
