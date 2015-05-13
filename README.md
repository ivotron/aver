# Validation Service

# Conventions

Each bucket is either an _independent_ or _dependent_ variable.

```
experiment_x.size:
experiment_x.method:
experiment_x.throughput:
```


# Example

```
for
  size > 3
expect
  ceph > raw
```
