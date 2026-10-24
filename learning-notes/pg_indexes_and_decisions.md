# PostgreSQL Indexes: How They Work and Why We Chose These Indexes

## 1. What is an index?

A database table stores the actual rows. An index is an additional
structure PostgreSQL maintains to help find rows efficiently.

For example, suppose `users` contains one million rows and we run:

``` sql
SELECT *
FROM users
WHERE username = 'dhwanil';
```

### Without an index

PostgreSQL may perform a sequential scan: it checks rows one by one,
testing whether each username matches. In the worst case, it examines
the whole table.

### With an index on `username`

When PostgreSQL builds the index, it records indexed values and
information that lets it locate the corresponding table rows.

For the query, PostgreSQL can:

1.  Search the index for `'dhwanil'`.
2.  Find the matching row location(s).
3.  Read the matching row(s) from the table to retrieve the requested
    columns.

This avoids checking every username. An index is like a book's index:
look up a term and jump to the relevant page instead of reading the
entire book.

A typical index lookup is much less work than scanning a large table,
but PostgreSQL chooses the plan it estimates will be cheapest. If a
query returns a large portion of the table, a sequential scan may be
faster.

Indexes also have costs: they use storage and need maintenance when
indexed data is inserted, updated, or deleted.

## 2. How PostgreSQL searches through an index

For an ordinary index, PostgreSQL commonly uses a B-tree. You do not
need to memorize its internal structure to understand the benefit.

The essential idea is that the index organizes searchable values so
PostgreSQL can narrow down where a value should be, instead of comparing
it against every table row.

This is similar to binary search in the sense that ordering helps narrow
the search. A database B-tree is not simply a binary-search operation
over a sorted copy of the table; it is a disk-friendly, balanced index
structure.

After finding an indexed key, PostgreSQL may still need to visit the
table to fetch columns that are not stored in the index. In some cases,
an index-only scan can answer a query from the index itself.

Use `EXPLAIN` to inspect the chosen plan:

``` sql
EXPLAIN
SELECT *
FROM posts
WHERE user_id = 10;
```

Use `EXPLAIN ANALYZE` to execute the query and show actual timing/row
counts:

``` sql
EXPLAIN ANALYZE
SELECT *
FROM posts
WHERE user_id = 10;
```

Look for nodes such as `Index Scan`, `Bitmap Index Scan`, or `Seq Scan`.
A sequential scan does not automatically mean the index is broken;
PostgreSQL may have estimated that scanning was cheaper.

> `EXPLAIN ANALYZE` executes the statement. Be careful when using it
> with statements that modify data.

## 3. Which columns are good candidates for indexes?

Index a column when your application often uses it to:

-   Find rows with `WHERE`.
-   Match rows in a `JOIN`.
-   Filter by a range or sort values, when the index type supports that
    operation.
-   Find referencing rows during certain foreign-key actions, such as
    deletes.

Selectivity matters: an index is often more useful when a search
condition narrows the result to a small subset of the table. For
example, a username is usually more selective than a boolean such as
`is_active`.

Do not index every column automatically. Each index consumes disk space
and adds work to writes.

## 4. Why these indexes are in our schema

The schema has indexes like:

``` sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_comments_content
    ON comments USING gin (content gin_trgm_ops);

CREATE INDEX idx_posts_title
    ON posts USING gin (title gin_trgm_ops);

CREATE INDEX idx_posts_tags
    ON posts USING gin (tags);

CREATE INDEX idx_users_username
    ON users (username);

CREATE INDEX idx_posts_user_id
    ON posts (user_id);

CREATE INDEX idx_comments_post_id
    ON comments (post_id);
```

### `users.username`

``` sql
CREATE INDEX idx_users_username ON users (username);
```

This helps lookups such as:

``` sql
SELECT *
FROM users
WHERE username = 'dhwanil';
```

The default index type here is B-tree. Usernames are a natural lookup
key when people search for or navigate to accounts by username.

If usernames must be unique, enforce that requirement with a `UNIQUE`
constraint (which creates a unique index), rather than relying on a
plain index:

``` sql
ALTER TABLE users
ADD CONSTRAINT users_username_unique UNIQUE (username);
```

A plain index improves lookup speed but does not prevent duplicate
usernames.

### `posts.user_id`

``` sql
CREATE INDEX idx_posts_user_id ON posts (user_id);
```

This helps queries such as:

``` sql
SELECT *
FROM posts
WHERE user_id = 10;
```

It also helps find posts belonging to a user during joins, for example:

``` sql
SELECT p.*
FROM users u
JOIN posts p ON p.user_id = u.id
WHERE u.id = 10;
```

`posts.id` does not need another ordinary index if it is already the
primary key: PostgreSQL automatically creates an index to enforce a
primary key. But that primary-key index is for finding a post by its own
ID. It does not efficiently answer "find all posts by user 10," which
filters on `user_id`.

### `comments.post_id`

``` sql
CREATE INDEX idx_comments_post_id ON comments (post_id);
```

This helps queries such as:

``` sql
SELECT *
FROM comments
WHERE post_id = 500;
```

It is useful for fetching a post's comments and for joins between posts
and comments.

`comments.id`, if it is the primary key, already has a primary-key
index. That index finds a comment by its own ID; it does not find all
comments attached to a particular post.

### Primary keys versus foreign-key columns

PostgreSQL automatically creates a unique index for a primary key (and
for a unique constraint). It does **not** automatically create an index
on the referencing/foreign-key column.

For example:

-   `users.id`: primary key, automatically indexed.
-   `posts.id`: primary key, automatically indexed.
-   `comments.id`: primary key, automatically indexed.
-   `posts.user_id`: foreign key referencing `users.id`; consider an
    index for queries that find a user's posts.
-   `comments.post_id`: foreign key referencing `posts.id`; consider an
    index for queries that find a post's comments.

The foreign-key indexes are chosen because of the access patterns, not
merely because the columns are foreign keys. They can also help
PostgreSQL locate dependent rows when a referenced row is deleted or
updated.

## 5. What is `pg_trgm`?

`pg_trgm` is a PostgreSQL extension that works with **trigrams**: groups
of three consecutive characters. It supports text similarity and can
help with substring searches.

For a simple illustration, the word `golang` contains these overlapping
three-character groups:

``` text
gol
ola
lan
ang
```

(Actual PostgreSQL trigram extraction also uses word-boundary padding
and has details beyond this simplified illustration.)

A trigram index can help with searches such as:

``` sql
SELECT *
FROM posts
WHERE title ILIKE '%golang%';
```

A regular B-tree usually cannot efficiently support a pattern that
starts with `%`, because the beginning of the searched value is unknown.
Trigrams give PostgreSQL smaller character pieces to search for.

Very short search patterns may not contain enough useful trigrams, so
the index may provide little benefit. PostgreSQL still checks the actual
search condition against candidate rows.

## 6. What is GIN, and what does "inverted index" mean?

GIN stands for **Generalized Inverted Index**. It is useful when a row
contains multiple searchable items, such as array elements, JSONB
components, or text trigrams.

Think of a forward mapping as:

``` text
Document/row → items it contains
```

An inverted mapping reverses that relationship:

``` text
Item → rows that contain it
```

For example, imagine posts with tags:

``` text
Post 1: ["go", "backend"]
Post 2: ["docker", "backend"]
Post 3: ["go", "database"]
```

Conceptually, an inverted index can answer:

``` text
"go"      → Post 1, Post 3
"backend" → Post 1, Post 2
"docker"  → Post 2
"database"→ Post 3
```

When searching for posts containing `go`, PostgreSQL can look up the
`go` entry and identify candidate rows rather than inspecting every
post's tag array.

For a PostgreSQL text array, GIN can support queries such as:

``` sql
-- Posts whose tags include "go"
SELECT *
FROM posts
WHERE tags @> ARRAY['go'];

-- Posts whose tags overlap with either "go" or "docker"
SELECT *
FROM posts
WHERE tags && ARRAY['go', 'docker'];
```

The supported operators depend on the column type and its GIN operator
class.

### GIN with trigrams

These indexes:

``` sql
CREATE INDEX idx_posts_title
    ON posts USING gin (title gin_trgm_ops);

CREATE INDEX idx_comments_content
    ON comments USING gin (content gin_trgm_ops);
```

tell PostgreSQL to use GIN with the trigram operator class for `title`
and `content`.

For a query such as:

``` sql
SELECT *
FROM comments
WHERE content ILIKE '%database%';
```

PostgreSQL can extract trigrams from the search pattern, use the index
to identify rows containing relevant trigrams, and then recheck the
original `ILIKE` condition. The index narrows the candidates; the
recheck confirms the full condition.

## 7. Trigram search versus full-text search

These are different approaches:

-   **Trigram search** is useful for character fragments, substring
    matching, and similarity. Example: find titles containing `lang`,
    even when it appears inside a larger word.
-   **Full-text search** is designed for word-oriented search. It
    converts text into normalized lexemes and supports matching words
    and combining search terms.

A basic PostgreSQL full-text index might be:

``` sql
CREATE INDEX idx_posts_title_fts
ON posts USING gin (to_tsvector('english', title));
```

And a matching query:

``` sql
SELECT *
FROM posts
WHERE to_tsvector('english', title)
      @@ plainto_tsquery('english', 'running backend');
```

Full-text search can normalize word forms (for example, language-aware
stemming); trigram search works with character sequences. Choose based
on what your search box should do.

## 8. Quick recap

  -----------------------------------------------------------------------
  Index                               Why it is useful
  ----------------------------------- -----------------------------------
  B-tree on `users.username`          Find a user by username

  Primary-key index on `users.id`     Find a user by ID; created
                                      automatically

  B-tree on `posts.user_id`           Find a user's posts and support
                                      related joins

  Primary-key index on `posts.id`     Find a post by ID; created
                                      automatically

  B-tree on `comments.post_id`        Find comments for a post and
                                      support related joins

  Primary-key index on `comments.id`  Find a comment by ID; created
                                      automatically

  GIN on `posts.tags`                 Find posts by tag elements (for
                                      supported operators/types)

  GIN with `gin_trgm_ops` on          Search title substrings or
  `posts.title`                       similarity

  GIN with `gin_trgm_ops` on          Search comment text substrings or
  `comments.content`                  similarity
  -----------------------------------------------------------------------

**Main takeaway:** indexes are extra lookup paths. Choose them based on
the queries your application commonly runs. Primary keys already have
indexes; additional indexes help find rows through other columns or
search mechanisms. GIN's inverted approach is especially helpful when
each row contains many searchable pieces, such as tags or trigrams.
