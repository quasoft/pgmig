# About pgmig

`pgmig` is a simple command line tool for continuous integration of changes to PostgreSQL databases. It's inspired by Martin Fowler's article "Evolutionary Database Design" and helps to apply the following practices:

- All database changes are migrations
- All database artifacts are version controlled with application code

If you need arguments on why those principles can be important, read the article.

The tool is written in Golang and once compiled can be run natively and without any dependencies on all platforms supported by the Go compiler.

## Conventions

Before using the tool, you have to adopt the following conventions:

- Create a separate folder under the project repository for database migration files (eg. `~/myproject/db/`)
- Create an SQL file named `00001_Create_database_structure.sql` that contains the initial structure of the database schema and common data (fixed tables like `countries`, `currencies`) as plain SQL
- Create such `.sql` files for all subsequent changes to the schema or common data with the following filename format:
  
      0000N_Description_of_the_change.sql
  
  where `N` is the unique version number of the migration (padded with zeroes) and `Description_of_the_change` is a human-readable description of the changes introduced by that SQL file.
  
  Example names of migration files:

      00001_Create_database_structure.sql
      00002_Populate_countries_table.sql
      00003_Add_contact_fields_to_person_table.sql
      00004_Add_function_counting_work_days_in_a_month.sql

## How to use
  
Create the necessary changelog table:

    pgmig init --host 10.0.0.1 -d testdb -U postgres

Check for pending migrations (which have not been applied to the database yet) in the PostgreSQL database specified by PG* environment variables:

    pgmig -D ~/myproject/db

Check for migrations by manually specifying connection details on the command line:

    pgmig -D ~/myproject/db --host 10.0.0.1 -d testdb -U postgres

Apply pending migrations:

    pgmig apply -D ~/myproject/db --host 10.0.0.1 -d testdb -U postgres