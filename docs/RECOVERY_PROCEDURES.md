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

## TV Shows not downloaded due to "failed to write file" error


List of all the processed_lines that failed to download due to "failed to write file" error.

```sql
SELECT p.id, d.id from public.download_info d
JOIN public.processed_lines p ON p.download_info_id = d.id
WHERE d.error_message like '%failed to write file%'
  AND state = 'failed'
```

**Save the results in a file** and then delete the records from the `processed_lines` table and the `download_info` table.

Delete the records from the `processed_lines` table.

```sql
DELETE FROM public.processed_lines WHERE id IN (
  SELECT p.id from public.download_info d
  JOIN public.processed_lines p ON p.download_info_id = d.id
  WHERE d.error_message like '%failed to write file%'
    AND state = 'failed'
)
```

As the records in the `processed_lines` table are deleted, the previous select/join query will return no results. So you have to delete the records from the `download_info` table based on the downloaded file 

```sql
DELETE FROM public.download_info d
WHERE d.id in (481, 482, ...)
```
