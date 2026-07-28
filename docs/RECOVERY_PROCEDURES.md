# Recovery procedure

## When tv shows are not well downloaded

Symptoms:

- Records in the table `tvshows` are present two times for the same show.
- Records in the table `processes_lines` only refer to one of the two records in `tvshows`.

Recovery steps:

- Clean up the `tvshows` table by removing duplicate records.

```sql
DELETE FROM public.tvshows t
-- WHERE tmdb_title like '%House of the Dragon%'
WHERE t.id > 1
AND NOT EXISTS (
	SELECT 1
	FROM public.processed_lines e
	WHERE e.tv_show_id = t.id
)
```
