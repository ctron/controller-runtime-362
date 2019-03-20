# Reproducer for kubernetes-sigs/controller-runtime#362

Log in to your OpenShift cluster.

Compile with:

    go build cmd/main

Run with:

    ./main

When you update dependencies run:

    glide install -v
