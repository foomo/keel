# NATS LOCAL TESTING

## SETUP

    $ docker network create nats-network
    $ docker run -d --name nats-main -p 4222:4222 -p 6222:6222 -p 8222:8222 --network=nats-network nats
    $ docker run -it --network=nats-network synadia/nats-box

## TESTS

Inside nats-box:

1 Publisher, 10 Million messages:

    # nats bench --pub=1 --msgs=1_0000000 --size=1 benchmarktest -s nats-main
    14:59:17 Starting benchmark [msgs=10,000,000, msgsize=1 B, pubs=1, subs=0]
        2s [==========================================================] 100%
    Pub stats: 4,943,149 msgs/sec ~ 4.71 MB/sec

1 Publisher, 10 subscribers, 10 Million messages:

    # nats bench --pub=1 --msgs=10000000 --size=1 --sub=10 benchmarktest -s nats-main
    15:00:58 Starting benchmark [msgs=10,000,000, msgsize=1 B, pubs=1, subs=10]
    20s [==========================================================] 100%

    NATS Pub/Sub stats: 5,212,646 msgs/sec ~ 4.97 MB/sec
    Pub stats: 473,983 msgs/sec ~ 462.87 KB/sec
    Sub stats: 4,739,128 msgs/sec ~ 4.52 MB/sec
    [1] 473,980 msgs/sec ~ 462.87 KB/sec (10000000 msgs)
    [2] 473,973 msgs/sec ~ 462.86 KB/sec (10000000 msgs)
    [3] 473,970 msgs/sec ~ 462.86 KB/sec (10000000 msgs)
    [4] 473,978 msgs/sec ~ 462.87 KB/sec (10000000 msgs)
    [5] 473,960 msgs/sec ~ 462.85 KB/sec (10000000 msgs)
    [6] 473,963 msgs/sec ~ 462.86 KB/sec (10000000 msgs)
    [7] 473,955 msgs/sec ~ 462.85 KB/sec (10000000 msgs)
    [8] 473,945 msgs/sec ~ 462.84 KB/sec (10000000 msgs)
    [9] 473,915 msgs/sec ~ 462.81 KB/sec (10000000 msgs)
    [10] 473,914 msgs/sec ~ 462.81 KB/sec (10000000 msgs)
    min 473,914 | avg 473,955 | max 473,980 | stddev 22 msgs

