![](logo.png)  
![](https://img.shields.io/badge/version-1.2.0-blue.svg)
![](https://img.shields.io/badge/LICENSE-AGPL-blue.svg)
[![Build Status](https://travis-ci.org/andyxning/shortme.svg?branch=master)](https://travis-ci.org/andyxning/shortme)
### Introduction
----
ShortMe is a url shortening service written in Golang.  
It is with high performance and scalable.  
ShortMe is ready to be used in production. Have fun with it. :)

### Features
----
* Convert same long urls to different short urls.
* Api support
* Web support
* Short url black list
    * To avoid some words, like `f**k` and `stupid`
    * To make sure that apis such as `/version` and `/health` will only be
    used as api not short urls or otherwise when requesting `http://127.0.0.1:3030/version`, version info will be returned rather the long url corresponding to the short url "version".
* Base string config in configuration file
    * **Once this base string is specified, it can not be reconfigured anymore
    otherwise the shortened urls may not be unique and thus may conflict with
     previous ones.**
* Avoid short url loop
    * In case we request the short url for an already shortened url by
    **shortme**. This is meaningless and will consume more resource in
    **shortme**.
* Short **http** or **https** urls

### Implementation
----
Currently, afaik, there are three ways to implement short url service.
* Hash
    * This way is straightforward. However, every hash function will have a
    collision when data is large.
* Sample
    * this way may contain collision, too. See example below (This example,
    in Python, is only used to demonstrate the collision situation.).

    ```python
    >>> import random
    >>> import string
    >>> random.sample('abc', 2)
    ['c', 'a']
    >>> random.sample('abc', 2)
    ['a', 'b']
    >>> random.sample('abc', 2)
    ['c', 'b']
    >>> random.sample('abc', 2)
    ['a', 'b']
    >>> random.sample('abc', 2)
    ['b', 'c']
    >>> random.sample('abc', 2)
    ['b', 'c']
    >>> random.sample('abc', 2)
    ['c', 'a']
    >>>
    ```
* Base
    * Just like converting bytes to base64 ascii, we can convert base10 to base62
    and then make a map between **0 .. 61** to **a-zA-Z0-9**. At last, we can
    get a unique string if we can make sure that the integer is unique.
    So, the URL shortening question transforms into making sure we can get a
    unique integer.
    ShortMe Use [the method that Flicker use](http://code.flickr.net/2010/02/08/ticket-servers-distributed-unique-primary-keys-on-the-cheap/)
    to generate a unique integer(Auto_increment + Replace into + MyISAM).
    Currently, we only use one backend db to generate sequence. For multiple
    sequence counter db configuration see [Deploy#Sequence Database]
    (#Sequence Database)

### Api
----
* `/version`
    * `HTTP GET`
    * Version info
    * Example
        * `curl http://127.0.0.1:3030/version`
* `/health`
    * `HTTP GET`
    * Health check
    * Example
        * `curl http://127.0.0.1:3030/health`
* `/short`
    * `HTTP POST`
    * Short the long url
    * Example
        * `curl -X POST -H "Content-Type:application/json" -d "{\"longURL\": \"http://www.google.com\"}" http://127.0.0.1:3030/short`
* `/{a-zA-Z0-9}{1,11}`
    * `HTTP GET`
    * Expand the short url and return a **temporary redirect** HTTP status
    * Example
        * `curl -v http://127.0.0.1:3030/3`

        ```bash
            *   Trying 127.0.0.1...
            * Connected to 127.0.0.1 (127.0.0.1) port 3030 (#0)
            > GET /3 HTTP/1.1
            > Host: 127.0.0.1:3030
            > User-Agent: curl/7.43.0
            > Accept: */*
            >
            < HTTP/1.1 307 Temporary Redirect
            < Location: http://www.google.com
            < Date: Fri, 15 Apr 2016 07:25:24 GMT
            < Content-Length: 0
            < Content-Type: text/plain; charset=utf-8
            <
            * Connection #0 to host 127.0.0.1 left intact
        ```

### Web
----
The web interface mainly used to make url shorting service more intuitively.

For **short** option, the shorted url, shorted url qr code and the 
corresponding long page is shown.

For **expand** option, the expanded url, expanded url qr code and the 
corresponding expanded page is shown. 

![](shortme_record.gif)


### Superstratum Projects
Projects that use `short-me`.

* [short-url](https://github.com/sillyhatxu/short-url)


### Install
----
#### Dependency
----
* Golang
* Mysql

#### Compile
----
```bash
mkdir -p $GOPATH/src/github.com/andyxning
cd $GOPATH/src/github.com/andyxning
git clone https://github.com/andyxning/shortme.git

cd shortme
make build
```

#### Database Schema
----
We use two databases. Import the two schemas.
* shortme
    * Store short url info
    * [shortme schema](schema/shortme.sql)
* sequence
    * sequence generator
    * [sequence schema](schema/sequence.sql)

#### Configuration
----
```
[http]
# Listen address
listen = "0.0.0.0:3030"

[sequence_db]
# Mysql sequence generator DSN
dsn = "sequence:sequence@tcp(127.0.0.1:3306)/sequence"

# Mysql connection pool max idle connection
max_idle_conns = 4

# Mysql connection pool max open connection
max_open_conns = 4

[short_db]
# Mysql short service read db DSN
read_dsn = "shortme_w:shortme_w@tcp(127.0.0.1:3306)/shortme"

# Mysql short service write db DSN
write_dsn = "shortme_r:shortme_r@tcp(127.0.0.1:3306)/shortme"

# Mysql connection pool max idle connection
max_idle_conns = 8

# Mysql connection pool max open connection
max_open_conns = 8

[common]
# short urls that will be filtered to use
black_short_urls = ["version","health","short","expand","css","js","fuck","stupid"]

# Base string used to generate short url
base_string = "Ds3K9ZNvWmHcakr1oPnxh4qpMEzAye8wX5IdJ2LFujUgtC07lOTb6GYBQViSfR"

# Short url service domain name. This is used to filter short url loop.
domain_name = "short.me:3030"

# Short url service schema: http or https.
schema = "http"
```
#### Capacity
----
We use an Mysql `unsigned bigint` type to store the sequence counter. According
 to the [Mysql doc](http://dev.mysql.com/doc/refman/5.7/en/integer-types.html)
 we can get `18446744073709551616` different integers.
 However, according to [Golang doc about `LastInsertId`](https://golang.org/pkg/database/sql/driver/#RowsAffected.LastInsertId)
 the returned auto increment integer can only be `int64` which will make the
 sequence smaller than `uint64`. Even through, we can still get
 `9223372036854775808` different integers and this will be large enough
 for most service.  

Supposing that  we consume `100,000,000` short urls one day, then the
sequence counter can last for `2 ** 63 / 100000000 / 365 = 252695124` years.

#### Short URL Length
----
The max string length needed for encoding `2 ** 63` integers will be **11**.

```python
>>> 62 ** 10
839299365868340224
>>> 2 ** 63
9223372036854775808L
>>> 62 ** 11
52036560683837093888L
```

#### Grant
----
After setting up the databases and before running **shortme**, make sure that
the corresponding user and password has been granted. After logging in mysql console, run following sql statement:
* `grant insert, delete on sequence.* to 'sequence'@'%' identified by 'sequence'`
* `grant insert on shortme.* to 'shortme_w'@'%' identified by 'shortme_w'`
* `grant select on shortme.* to 'shortme_r'@'%' identified by 'shortme_r'`

#### Run
----
* make sure that `static` directory will be at the same directory as **shortme**
* `./shortme -c config.conf`

### Deploy
----

#### <a name="Sequence Database"></a>Sequence Database
----
In the [Flickr blog](http://code.flickr.net/2010/02/08/ticket-servers-distributed-unique-primary-keys-on-the-cheap/),
Flickr suggests that we can use two databases with one for even sequence and
the other one for odd sequence. This will make sequence generator being more
available in case one database is down and will also spread the load about
generate sequence. After splitting sequence db from one to more, we can use
[HaProxy](http://www.haproxy.org/) as a reverse proxy and thus more sequence
databases can be used as one. As for load balance algorithm, i think **round
robin** is good enough for this situation.

In two databases situation, we should add the following configuration to each
 database configuration file.
* First database

```
auto_increment_offset 1
auto_increment_increment 2
```

* Second databse

```
auto_increment_offset 2
auto_increment_increment 2
```

Then each time to generate a sequence counter, we can execute below sql
statement:  
`replace into sequence(stub) values("sequence")`

In cases we use three databases as sequence counter generator, we should
insert a record for each table in two databases.
* First database

```
auto_increment_offset 1
auto_increment_increment 3
```

* Second database

```
auto_increment_offset 2
auto_increment_increment 3
```

* Third database

```
auto_increment_offset 3
auto_increment_increment 3
```

Then each time to generate a sequence counter, we can execute below sql
statement:  
`replace into sequence(stub) values("sequence")`

Ok, i think you get the point. When using `N` databases to generate sequence
counter, configuration for each database configuration file will just
like below:

```
for i := range N {
    add "auto_increment_offset i" to config file
    add "auto_increment_increment N" to config file
}

```
So, sequence generator can be horizontally scalable.

#### Shard
----
With short urls increasing, many records are stored in one table. This
 is not an optimal mysql practice. In this case we can simply shard table to
 bypass this problem.

For example, we can shard according to the **base integer** using **modula hash
 algorithm**. This has a good distribution between tables. We can use **100**
 **short** tables with names like **short_00/short_01/short_02/..
 ./short_99**. we can use pseudo code blow to determine which is the
 table to store the short url record.

 ```
 baseInteger := sequence.NextSequence()
 tableName := fmt.Sprintf("short_%s", baseInteger % 100)
 ```

 There are many table sharding algorithms, we can shard table according to
 range id, user name and so on. If we use user name as the criteria to shard
 table, we can do some aggregate algorithm like how many records a user has
 created easily. This may also has some drawbacks such as if user **Lily**
 and user **Lucy** are sharded to different tables and **Lily** shorts about
 **1k** urls **Lucy** shorts about **1M** urls, then we may encounter the
 unbalance hash problem, i.e., some tables contains more records than others.

In conclusion, there are many factors to consider before we can make a
decision which hash algorithm to use.

#### Statistics
----
Sometimes we may want to make some statistics about hit number, UA(User 
Agent), original IP and so on. 

A recommended way to deploy **shortme** is to use it behind a reverse proxy 
server such as **Nginx**. Under this way, the statistics info can be analysed
 by analysing the access log of **Nginx**. This is can be accomplished by 
 `awk` or more trending log analyse stack [`ELK`](https://www.elastic.co/).

### Problems
----
* long url may make the generated qr code unreadable. I have test this in my 
self phone. This remains to be tested more meticulous.  
* One demand about customizing the short url can not be done easily currently
 in **shortme** according to the id generation logic. Let's make it happen. :)
