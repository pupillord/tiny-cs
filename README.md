**Tiny Client/Server**

a tiny client and server for tinysql.

**Start**

```
// start server 
cd sql-server
go run .

// start client 
cd sql-client
go run .

// run some sql in client
create table student(int id, varchar name);
select * from student;
```

**Note**
* There is no data in server, so you will only get "the query has been completed" for any sql query.
* The result in server are defined in func `handleQuery` of `conn.go`.
* The rules of protocol in Client/Server can be found in `protocol.go`