# The technical task for backend position

Fork repository to solve the task.

## Description

We have applications from customers in next format (one user can have multiple applications):

```javascript
{
    id: String; // mongo object Id
    Status: String; // One of "Open"/"In progress"/"Closed"; Open - initial status
    UserId: String; // mongo object Id
    CreatedAt: String; // time.RFC3339
    UpdatedAt: String; // time.RFC3339 = when application was updated last time
    ExternalStatus: String; // "Processed"/"Skipped" - this status from external system;
}
```

You need create backend system which will be next requests:

```proto

// Create application, should return application
message CreateApplicationRequest {
  string user_id = 1;
}

// Should return single application, should return application
message GetByIdRequest {
  string id = 1;
}

// You should create by you self
// Should return array of applications
message GetByFiltersRequest {
  /* Here should be:
    1. Status
    2. CreatedAt time range
    3. UpdatedAt time range
    4. UserId
  */
}

// You should create by you self, should return application
message UpdateApplication {
  // You can update just status +  internal update updatedAt
}
```

## Technical requirements

1. Application should use gRPC framework for implement messages + server
2. Should use mongo
3. Should have cache for get requests (GetByIdRequest,GetByFiltersRequest) + invalidate cache on update

Additional:
1. Implement github ci for build docker image
2. Add unit/integration tests
3. Create docker-compose.yml for run it locally
4. Makefile + linters

Try to create optimal mongo schema + protobuf schemas

## External system

**Important** - external system your not control from code, status can be changed any time (you dont know when). So you can create different assumption for caching.

`ExternalStatus` - it is status from external system. Status processed on flight. 

Mock for it:
```go   
// http://localhost:4200/status/60119e16a4e9e747878c8887 - get status
external.GetEngine().Run(":4200")
```
