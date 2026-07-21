-- Seed Test Data for Downloads Enrichment Feature
-- This script creates realistic test data for validating the enriched downloads display

-- Enable required PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Clean existing test data (optional - comment out to preserve existing data)
-- DELETE FROM download_info WHERE id > 0;
-- DELETE FROM processed_lines WHERE id > 0;
-- DELETE FROM movies WHERE id > 0;
-- DELETE FROM tvshows WHERE id > 0;

-- Reset sequences to start from a known point
-- SELECT setval('download_info_id_seq', 1, false);
-- SELECT setval('processed_lines_id_seq', 1, false);
-- SELECT setval('movies_id_seq', 1, false);
-- SELECT setval('tvshows_id_seq', 1, false);

-- =============================================================================
-- MOVIES - Complete downloads with various quality and year scenarios
-- =============================================================================

-- Movie 1: La Cité de Dieu (2002) - Perfect case
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (598, 'La Cité de Dieu', 2002, '"Crime", "Drama"', 130, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/city-of-god.m3u8',
    'completed',
    '/media/movies/La.Cite.de.Dieu.2002.1080p.BluRay.x264/La.Cite.de.Dieu.2002.1080p.BluRay.x264.mkv',
    4516241408,
    4516241408,
    4516241408,
    0,
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '3 hours',
    NOW() - INTERVAL '1 hour'
) RETURNING id AS download_id_1;

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="City of God" group-title="VOD Movies"',
    'https://provider.com/stream/city-of-god.m3u8',
    encode(digest('city-of-god-2002', 'sha256'), 'hex'),
    'City of God',
    'VOD Movies',
    NOW() - INTERVAL '3 hours',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = 'La Cité de Dieu' AND tmdb_year = 2002),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '3 hours',
    NOW() - INTERVAL '1 hour',
    '1080p'
);

-- Movie 2: Matrix (1999) - 4K quality
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (603, 'The Matrix', 1999, '"Action", "Science Fiction"', 136, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/matrix.m3u8',
    'completed',
    '/media/movies/The.Matrix.1999.2160p.UHD.BluRay.x265/The.Matrix.1999.2160p.mkv',
    8589934592,
    8589934592,
    8589934592,
    0,
    NOW() - INTERVAL '5 hours',
    NOW() - INTERVAL '3 hours',
    NOW() - INTERVAL '6 hours',
    NOW() - INTERVAL '3 hours'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="The Matrix" group-title="VOD Movies 4K"',
    'https://provider.com/stream/matrix.m3u8',
    encode(digest('matrix-1999-4k', 'sha256'), 'hex'),
    'The Matrix',
    'VOD Movies 4K',
    NOW() - INTERVAL '6 hours',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = 'The Matrix' AND tmdb_year = 1999),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '6 hours',
    NOW() - INTERVAL '3 hours',
    '2160p'
);

-- Movie 3: Avatar (2009) - Missing year in path (problem case)
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (19995, 'Avatar', 2009, '"Action", "Adventure", "Fantasy", "Science Fiction"', 162, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/avatar.m3u8',
    'completed',
    '/media/movies/Avatar.BluRay.1080p.x264/avatar.mkv',
    7318349394,
    7318349394,
    7318349394,
    0,
    NOW() - INTERVAL '1 day',
    NOW() - INTERVAL '23 hours',
    NOW() - INTERVAL '1 day 1 hour',
    NOW() - INTERVAL '23 hours'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="Avatar" group-title="VOD Movies"',
    'https://provider.com/stream/avatar.m3u8',
    encode(digest('avatar-2009', 'sha256'), 'hex'),
    'Avatar',
    'VOD Movies',
    NOW() - INTERVAL '1 day 1 hour',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = 'Avatar' AND tmdb_year = 2009),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '1 day 1 hour',
    NOW() - INTERVAL '23 hours',
    '1080p'
);

-- Movie 4: Inception (2010) - Year mismatch (wrong year in path)
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (27205, 'Inception', 2010, '"Action", "Science Fiction", "Adventure"', 148, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/inception.m3u8',
    'completed',
    '/media/movies/Inception.2011.720p.BluRay/inception.mp4',
    3221225472,
    3221225472,
    3221225472,
    0,
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days' + INTERVAL '2 hours',
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days' + INTERVAL '2 hours'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="Inception" group-title="VOD Movies"',
    'https://provider.com/stream/inception.m3u8',
    encode(digest('inception-2010', 'sha256'), 'hex'),
    'Inception',
    'VOD Movies',
    NOW() - INTERVAL '2 days',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = 'Inception' AND tmdb_year = 2010),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '2 days',
    NOW() - INTERVAL '2 days' + INTERVAL '2 hours',
    '720p'
);

-- Movie 5: Pulp Fiction (1994) - Low quality (480p)
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (680, 'Pulp Fiction', 1994, '"Thriller", "Crime"', 154, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/pulp-fiction.m3u8',
    'completed',
    '/media/movies/Pulp.Fiction.1994.480p.DVDRip/pulp.fiction.avi',
    734003200,
    734003200,
    734003200,
    0,
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '3 days' + INTERVAL '30 minutes',
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '3 days' + INTERVAL '30 minutes'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="Pulp Fiction" group-title="VOD Movies Classic"',
    'https://provider.com/stream/pulp-fiction.m3u8',
    encode(digest('pulp-fiction-1994', 'sha256'), 'hex'),
    'Pulp Fiction',
    'VOD Movies Classic',
    NOW() - INTERVAL '3 days',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = 'Pulp Fiction' AND tmdb_year = 1994),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '3 days',
    NOW() - INTERVAL '3 days' + INTERVAL '30 minutes',
    '480p'
);

-- =============================================================================
-- TV SHOWS - Various states and qualities
-- =============================================================================

-- TV Show 1: Breaking Bad S05E14 - Perfect case
INSERT INTO tvshows (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, season, episode, created_at, updated_at)
VALUES (1396, 'Breaking Bad', 2008, '"Drama", "Crime"', 5, 14, NOW(), NOW());

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/breaking-bad-s05e14.m3u8',
    'completed',
    '/media/tvshows/Breaking.Bad/Season.05/Breaking.Bad.S05E14.1080p.WEB-DL.mkv',
    1932735283,
    1932735283,
    1932735283,
    0,
    NOW() - INTERVAL '12 hours',
    NOW() - INTERVAL '11 hours',
    NOW() - INTERVAL '13 hours',
    NOW() - INTERVAL '11 hours'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, tv_show_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="Breaking Bad S05E14" group-title="VOD Series"',
    'https://provider.com/stream/breaking-bad-s05e14.m3u8',
    encode(digest('breaking-bad-s05e14', 'sha256'), 'hex'),
    'Breaking Bad S05E14',
    'VOD Series',
    NOW() - INTERVAL '13 hours',
    'tvshows',
    (SELECT id FROM tvshows WHERE tmdb_title = 'Breaking Bad' AND season = 5 AND episode = 14),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '13 hours',
    NOW() - INTERVAL '11 hours',
    '1080p'
);

-- TV Show 2: Game of Thrones S08E06 - 4K
INSERT INTO tvshows (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, season, episode, created_at, updated_at)
VALUES (1399, 'Game of Thrones', 2011, '"Sci-Fi & Fantasy", "Drama", "Action & Adventure"', 8, 6, NOW(), NOW());

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/got-s08e06.m3u8',
    'completed',
    '/media/tvshows/Game.of.Thrones/Season.08/Game.of.Thrones.S08E06.2160p.UHD.mkv',
    5368709120,
    5368709120,
    5368709120,
    0,
    NOW() - INTERVAL '1 day 6 hours',
    NOW() - INTERVAL '1 day 4 hours',
    NOW() - INTERVAL '1 day 7 hours',
    NOW() - INTERVAL '1 day 4 hours'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, tv_show_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="Game of Thrones S08E06" group-title="VOD Series 4K"',
    'https://provider.com/stream/got-s08e06.m3u8',
    encode(digest('got-s08e06-4k', 'sha256'), 'hex'),
    'Game of Thrones S08E06 Finale',
    'VOD Series 4K',
    NOW() - INTERVAL '1 day 7 hours',
    'tvshows',
    (SELECT id FROM tvshows WHERE tmdb_title = 'Game of Thrones' AND season = 8 AND episode = 6),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '1 day 7 hours',
    NOW() - INTERVAL '1 day 4 hours',
    '2160p'
);

-- =============================================================================
-- FAILED DOWNLOADS
-- =============================================================================

-- Failed Movie: The Dark Knight (2008) - Failed with error
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (155, 'The Dark Knight', 2008, '"Drama", "Action", "Crime", "Thriller"', 152, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, error_message, started_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/dark-knight.m3u8',
    'failed',
    NULL,
    NULL,
    1073741824,
    4294967296,
    3,
    'HTTP 403 Forbidden - Link expired',
    NOW() - INTERVAL '6 hours',
    NOW() - INTERVAL '7 hours',
    NOW() - INTERVAL '5 hours'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="The Dark Knight" group-title="VOD Movies"',
    'https://provider.com/stream/dark-knight.m3u8',
    encode(digest('dark-knight-2008', 'sha256'), 'hex'),
    'The Dark Knight',
    'VOD Movies',
    NOW() - INTERVAL '7 hours',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = 'The Dark Knight' AND tmdb_year = 2008),
    (SELECT currval('download_info_id_seq')),
    'failed',
    NOW() - INTERVAL '7 hours',
    NOW() - INTERVAL '5 hours',
    '1080p'
);

-- Failed TV Show: The Office S02E01 - Network error
INSERT INTO tvshows (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, season, episode, created_at, updated_at)
VALUES (2316, 'The Office', 2005, '"Comedy"', 2, 1, NOW(), NOW());

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, error_message, started_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/office-s02e01.m3u8',
    'failed',
    NULL,
    NULL,
    524288000,
    1048576000,
    5,
    'Network timeout - Connection reset by peer',
    NOW() - INTERVAL '4 hours',
    NOW() - INTERVAL '5 hours',
    NOW() - INTERVAL '3 hours 30 minutes'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, tv_show_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="The Office S02E01" group-title="VOD Comedy"',
    'https://provider.com/stream/office-s02e01.m3u8',
    encode(digest('office-s02e01', 'sha256'), 'hex'),
    'The Office S02E01 - The Dundies',
    'VOD Comedy',
    NOW() - INTERVAL '5 hours',
    'tvshows',
    (SELECT id FROM tvshows WHERE tmdb_title = 'The Office' AND season = 2 AND episode = 1),
    (SELECT currval('download_info_id_seq')),
    'failed',
    NOW() - INTERVAL '5 hours',
    NOW() - INTERVAL '3 hours 30 minutes',
    '720p'
);

-- =============================================================================
-- IN-PROGRESS DOWNLOADS
-- =============================================================================

-- Downloading: Interstellar (2014)
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (157336, 'Interstellar', 2014, '"Adventure", "Drama", "Science Fiction"', 169, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, locked_at, locked_by, started_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/interstellar.m3u8',
    'downloading',
    '/media/movies/Interstellar.2014.1080p.BluRay.x264/interstellar.mkv.part',
    NULL,
    2684354560,
    6442450944,
    0,
    NOW() - INTERVAL '30 minutes',
    'stalkeer-worker-1',
    NOW() - INTERVAL '45 minutes',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '1 minute'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="Interstellar" group-title="VOD Movies"',
    'https://provider.com/stream/interstellar.m3u8',
    encode(digest('interstellar-2014', 'sha256'), 'hex'),
    'Interstellar',
    'VOD Movies',
    NOW() - INTERVAL '1 hour',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = 'Interstellar' AND tmdb_year = 2014),
    (SELECT currval('download_info_id_seq')),
    'downloading',
    NOW() - INTERVAL '1 hour',
    NOW() - INTERVAL '1 minute',
    '1080p'
);

-- Retrying: Stranger Things S04E09
INSERT INTO tvshows (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, season, episode, created_at, updated_at)
VALUES (66732, 'Stranger Things', 2016, '"Sci-Fi & Fantasy", "Mystery", "Drama"', 4, 9, NOW(), NOW());

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, last_retry_at, started_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/stranger-things-s04e09.m3u8',
    'retrying',
    '/media/tvshows/Stranger.Things/Season.04/Stranger.Things.S04E09.720p.WEB.mkv.part',
    NULL,
    897581056,
    1610612736,
    2,
    NOW() - INTERVAL '5 minutes',
    NOW() - INTERVAL '2 hours',
    NOW() - INTERVAL '2 hours 30 minutes',
    NOW() - INTERVAL '5 minutes'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, tv_show_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="Stranger Things S04E09" group-title="VOD Series"',
    'https://provider.com/stream/stranger-things-s04e09.m3u8',
    encode(digest('stranger-things-s04e09', 'sha256'), 'hex'),
    'Stranger Things S04E09 - The Piggyback',
    'VOD Series',
    NOW() - INTERVAL '2 hours 30 minutes',
    'tvshows',
    (SELECT id FROM tvshows WHERE tmdb_title = 'Stranger Things' AND season = 4 AND episode = 9),
    (SELECT currval('download_info_id_seq')),
    'downloading',
    NOW() - INTERVAL '2 hours 30 minutes',
    NOW() - INTERVAL '5 minutes',
    '720p'
);

-- =============================================================================
-- EDGE CASES
-- =============================================================================

-- Unknown format (.flv)
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (13, 'Forrest Gump', 1994, '"Comedy", "Drama", "Romance"', 142, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/forrest-gump.m3u8',
    'completed',
    '/media/movies/Forrest.Gump.1994.DVDRip/forrest.gump.flv',
    1073741824,
    1073741824,
    1073741824,
    0,
    NOW() - INTERVAL '4 days',
    NOW() - INTERVAL '4 days' + INTERVAL '1 hour',
    NOW() - INTERVAL '4 days 1 hour',
    NOW() - INTERVAL '4 days' + INTERVAL '1 hour'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at)
VALUES (
    '#EXTINF:-1 tvg-name="Forrest Gump" group-title="VOD Classic Movies"',
    'https://provider.com/stream/forrest-gump.m3u8',
    encode(digest('forrest-gump-1994', 'sha256'), 'hex'),
    'Forrest Gump',
    'VOD Classic Movies',
    NOW() - INTERVAL '4 days 1 hour',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = 'Forrest Gump' AND tmdb_year = 1994),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '4 days 1 hour',
    NOW() - INTERVAL '4 days' + INTERVAL '1 hour',
    NULL
);

-- Ambiguous year: 2001 A Space Odyssey (1968)
INSERT INTO movies (tmdb_id, tmdb_title, tmdb_year, tmdb_genres, duration, created_at, updated_at)
VALUES (62, '2001: A Space Odyssey', 1968, '"Science Fiction", "Mystery", "Adventure"', 149, NOW(), NOW())
ON CONFLICT (tmdb_title, tmdb_year) DO NOTHING;

INSERT INTO download_info (url, status, download_path, file_size, bytes_downloaded, total_bytes, retry_count, started_at, completed_at, created_at, updated_at)
VALUES (
    'https://provider.com/stream/2001-space-odyssey.m3u8',
    'completed',
    '/media/movies/2001.A.Space.Odyssey.1968.720p.BluRay/2001.space.odyssey.mkv',
    3758096384,
    3758096384,
    3758096384,
    0,
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '5 days' + INTERVAL '1 hour 30 minutes',
    NOW() - INTERVAL '5 days 1 hour',
    NOW() - INTERVAL '5 days' + INTERVAL '1 hour 30 minutes'
);

INSERT INTO processed_lines (line_content, line_url, line_hash, tvg_name, group_title, processed_at, content_type, movie_id, download_info_id, state, created_at, updated_at, resolution)
VALUES (
    '#EXTINF:-1 tvg-name="2001: A Space Odyssey" group-title="VOD Sci-Fi Classics"',
    'https://provider.com/stream/2001-space-odyssey.m3u8',
    encode(digest('2001-space-odyssey-1968', 'sha256'), 'hex'),
    '2001: A Space Odyssey',
    'VOD Sci-Fi Classics',
    NOW() - INTERVAL '5 days 1 hour',
    'movies',
    (SELECT id FROM movies WHERE tmdb_title = '2001: A Space Odyssey' AND tmdb_year = 1968),
    (SELECT currval('download_info_id_seq')),
    'downloaded',
    NOW() - INTERVAL '5 days 1 hour',
    NOW() - INTERVAL '5 days' + INTERVAL '1 hour 30 minutes',
    '720p'
);

-- Summary
SELECT 
    'Test data seed complete!' as message,
    (SELECT COUNT(*) FROM movies) as movies_count,
    (SELECT COUNT(*) FROM tvshows) as tvshows_count,
    (SELECT COUNT(*) FROM download_info) as downloads_count,
    (SELECT COUNT(*) FROM processed_lines) as processed_lines_count,
    (SELECT COUNT(*) FROM download_info WHERE status = 'completed') as completed_downloads,
    (SELECT COUNT(*) FROM download_info WHERE status = 'failed') as failed_downloads,
    (SELECT COUNT(*) FROM download_info WHERE status IN ('downloading', 'retrying')) as active_downloads;
