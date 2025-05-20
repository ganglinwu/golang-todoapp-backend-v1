- mongoDB is a nightmare to work with if we have nested objects

consider the following schema

Proj = {
    _id: ObjectID('somehexstring'),
    name: "proj 1",
    tasks : [
      {
          _id: ObjectID('hexstring'),
          name: "task 1",
          description: "description 1",
          duedate: 2025-01-01T00:00:00.000+00:00,
          priority: "low",
        },
        {nested object 2 that i'm lazy to flesh out},
    ]
  }

  the context of this example is a todo app that allows you to categorize todos under projects

  imagine if you want to query/find a specific todo(task) only

  you would have to create a 3 stage aggregation pipeline
   - match
   - replace root
   - (set) projection/filter



  I finally start to understand why people prefer dynamoDB or even postgreSQL over mongoDB

  my guess is if the schema is even more complicated.. the json overhead is just going to be hard to "unwind"

  query speeds will also suffer if we scale up


  - mongoDB has upsert! (it will insert new entry when cannot find a matching update!)

  firstly most of mongoDB's update/replace methods have the following parameters

    e.g. UpdateByID( <golang context>, <query filter>, <update>, <options>)

  to enable upsert we will have to set 
    opts := options.UpdateByID().SetUpsert(true)



  - BUT upsert uses the same operator($) as arrayFilters!
    - that means you cannot use upsert when you want to update documents in an array!
  ```
  From mongoDB DOCs (https://www.mongodb.com/docs/manual/reference/operator/update/positional/)

  Do not use the positional operator $ with upsert operations because inserts will use the $ as a field name in the inserted document.
  ```
