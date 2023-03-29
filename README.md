# Aggregator Service

This repo contains code for an aggregator service that 
1. Takes the discovery URL, i.e. `http://localhost:8080/shops` as a command line argument.
2. Queries the discovery URL to find out all the shops and their flavor URLs.
3. Queries each of these per shop flavor URLs, i.e. `http://localhost:8080/shops/shop1/flavors`, and aggregates them.
    * Multiple shops can have the same flavor, but the aggregate list should be unique. So, if 2 shops report the `rum raisin` flavor,
      there should only be a single entry in the aggregate.
4. Exposes a local http server with a single endpoint of `/flavors`, which returns the aggregated list of flavors across all ice cream shops.
5. Keep this aggregated list of flavors up-to-date by repeating steps 2 and 3 every few seconds.

We have a discovery endpoint that provides us a list of ice cream shop names and URLs for flavors available at that shop:

```
$ curl http://localhost:8080/shops 
shop1,http://localhost:8080/shops/shop1/flavors
shop2,http://localhost:8080/shops/shop2/flavors
shop3,http://localhost:8080/shops/shop3/flavors
```

Each flavor URL returns a list of flavors available at that shop:

```
$ curl http://localhost:8080/shops/shop1/flavors
rum raisin
cherry garcia
```

We want to provide an aggregated endpoint to our customers which tells them all the flavors available across all the shops:

```
$ curl http://localhost:8081/flavors
rola cola
rum raisin
chunky monkey
mcflurry
cherry garcia
```

# Constraints

* The aggregated `/flavors` endpoint should respond in less than 1s.
* The shops can be added or removed any time. The aggregator should stay up-to-date with the current list of shops within 1 minute of the list changing from the `/shops` url.
* The `/shops` url can take upto 10s to respond and can return 1-10K shops.
* The flavors for each shop can also be added or removed any time. The aggregator should stay up-to-date with the current list of flavors within 3 minutes of the flavor list changing for a shop.
* One shop can have 1-1000 flavors and can take upto 10s to respond to requests for listing its flavors.
* It is acceptable to return a partial list of flavors when the application is initializing.
