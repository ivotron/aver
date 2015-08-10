# Project Aver

[![Build Status](https://travis-ci.org/ivotron/aver.svg?branch=master)](https://travis-ci.org/ivotron/aver)

Aver lets you quickly test claims made about the performance of a 
system. Aver can be thought as a system that checks high-level 
assertions (in the 
[TDD](http://en.wikipedia.org/wiki/Test-driven_development) sense) 
against a dataset corresponding to system performance metrics.

## Overview

Consider the following gathered metrics:

|Size|Workload|Method|Replication|Throughput|
|:--:|:------:|:----:|:---------:|:--------:|
|  1 | read   |   a  |    1      |    141   |
|  1 | read   |   b  |    1      |    145   |
|  1 | write  |   a  |    1      |    142   |
|  1 | write  |   b  |    1      |    149   |
|  . |  ...   |   .  |    .      |    .     |
|  . |  ...   |   .  |    .      |    .     |
| 32 | write  |   a  |    5      |    210   |
| 32 | write  |   b  |    5      |    136   |

The above (truncated) table contains performance measurements of a 
distributed storage system (throughput column in MB/s). The throughput 
is in function of the size of the cluster, replication factor (how 
many copies of the data), type of request (read/write) and the method 
used to replicate data.

Now, suppose we make the claim that method `a` beats method `b` by 2x. 
At this point, you have two options: (1) you believe in our expertise 
and trust our word, or (2) you re-run the test and check the results 
to confirm our claims. Aver shortens the time it takes to validate 
results by providing a simple language that can be used to express 
this type of assertions, as well as a small program that test the 
validations over the captured data. The following statement expresses 
our claim:

```
for
  size > 4 and replication = *
expect
  throughput(method='a') > throughput(method='b') * 2
```

In prose, the above states that, regardless of the replication factor, 
method `a` is twice as fast as method `b` when the size of the cluster 
goes above a certain threshold (4 in this case). Given this statement 
and a pointer to where the dataset resides, aver checks whether this 
validation holds.

## Usage

There are two ways of using Aver. Programatically or through the CLI. 

<!--
Using the CLI, the following would check the validation statement made 
earlier:

```sh
aver --host=localhost --metrics=
```

Assuming results are stored in one of the supported backends (see 
Backends section below), the above checks the validation clause 
against. Check the API documentation to learn how to use Aver 
programatically.

## Motivation

Making it easier to corroborate experimental results in CS research. 
Aver provides a mechanism to express and test hypothesis about the 
behavior of a system.

## Validation Language

### Syntax

## Backends

## Installation

## Internals
-->
