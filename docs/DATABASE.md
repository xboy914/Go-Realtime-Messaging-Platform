# PostgreSQL model

The database owns durable users, rooms, memberships, and messages. SQL migrations are embedded in
the server binary, ordered by filename, tracked in `schema_migrations`, and applied transactionally.

A message insert uses `INSERT ... SELECT ... WHERE EXISTS` so membership authorization and creation
are one database statement. History is indexed by room and reverse chronology, limited to 50 records,
then returned oldest-first for rendering.

The included identities and room are synthetic demo fixtures. No production data is present.
