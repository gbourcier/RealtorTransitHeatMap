-- Local dev seed data: 10 listings around Montréal, 30 daily price-history rows each.
-- Run with:  psql "$DATABASE_URL" -f seeds/seed.sql
-- Re-running is safe: TRUNCATE wipes existing rows first.

TRUNCATE listings, listing_price_history RESTART IDENTITY CASCADE;

INSERT INTO listings (board, mls, latitude, longitude, address, photo_url, is_available,
                        building_type, bedroom_count, bathroom_count, interior_area_sqft,
                        commute_seconds_downtown, commute_computed_at,
                        first_seen_at, last_updated_at)
VALUES
    (1234, 12345678, 45.5240, -73.5810, '1234 Rue Mont-Royal E, Montréal',      'https://cdn.realtor.ca/listings/TS639139196971800000/reb5/highres/0/18465520_1.jpg', TRUE,  1, 3, 2, 1400,  600, now(), now() - interval '30 days', now()),
    (1234, 12345679, 45.4860, -73.5910, '456 Av. Westmount, Westmount',          'https://cdn.realtor.ca/listings/TS639139196398330000/reb5/highres/8/17178928_1.jpg', TRUE,  1, 4, 3, 2200,  900, now(), now() - interval '30 days', now()),
    (1234, 12345680, 45.5200, -73.6050, '789 Av. Bernard, Outremont',            'https://cdn.realtor.ca/listings/TS639139133354170000/reb5/highres/4/13073744_1.jpg', TRUE,  1, 3, 2, 1550, 1200, now(), now() - interval '30 days', now()),
    (1234, 12345681, 45.4580, -73.5670, '234 Rue Wellington, Verdun',            'https://cdn.realtor.ca/listings/TS639138833382700000/reb5/highres/0/11956400_1.jpg', TRUE,  1, 2, 1,  950, 1500, now(), now() - interval '30 days', now()),
    (1234, 12345682, 45.4700, -73.6200, '567 Rue Sherbrooke O, Montréal',        'https://cdn.realtor.ca/listings/TS639138767309270000/reb5/highres/7/25160737_1.jpg', FALSE, 1, 3, 2, 1300, 1800, now(), now() - interval '30 days', now()),
    (1234, 12345683, 45.5240, -73.6010, '890 Av. Laurier, Montréal',             'https://cdn.realtor.ca/listings/TS639138767267100000/reb5/highres/8/25037898_1.jpg', TRUE,  1, 3, 2, 1450, 1100, now(), now() - interval '30 days', now()),
    (1234, 12345684, 45.5460, -73.5850, '321 Rue Beaubien, Montréal',            'https://cdn.realtor.ca/listings/TS639138764422770000/reb5/highres/5/16404265_1.jpg', FALSE, 1, 2, 1, 1000, 1400, now(), now() - interval '30 days', now()),
    (1234, 12345685, 45.4770, -73.5850, '654 Rue Notre-Dame O, Montréal',        'https://cdn.realtor.ca/listings/TS639138763165370000/reb5/highres/1/13474951_1.jpg', TRUE,  1, 3, 2, 1250, 1000, now(), now() - interval '30 days', now()),
    (1234, 12345686, 45.5410, -73.5470, '987 Rue Ontario E, Montréal',           'https://cdn.realtor.ca/listings/TS639138762229130000/reb5/highres/7/11567797_1.jpg', TRUE,  1, 2, 1,  900, 1300, now(), now() - interval '30 days', now()),
    (1234, 12345687, 45.5050, -73.5670, '1010 Rue Sainte-Catherine O, Montréal', 'https://cdn.realtor.ca/listings/TS639138694681870000/reb5/highres/4/23434264_1.jpg', TRUE,  1, 4, 2, 1800,  300, now(), now() - interval '30 days', now());

-- 30 history rows per listing, spaced 1 day apart, with a small downward drift
-- over time (today's price is ~0.9% below the price 30 days ago — i.e. mild
-- "price reduction" pattern that's typical for active listings).
INSERT INTO listing_price_history (board, mls, observed_at, price)
SELECT
    p.board,
    p.mls,
    date_trunc('day', now()) - (gs * interval '1 day'),
    ROUND((p.base_price * (1 + 0.0003 * gs))::numeric, 2)
FROM (VALUES
    (1234, 12345678,  625000.00),
    (1234, 12345679, 1150000.00),
    (1234, 12345680,  895000.00),
    (1234, 12345681,  475000.00),
    (1234, 12345682,  735000.00),
    (1234, 12345683,  815000.00),
    (1234, 12345684,  560000.00),
    (1234, 12345685,  695000.00),
    (1234, 12345686,  415000.00),
    (1234, 12345687,  980000.00)
) AS p(board, mls, base_price)
CROSS JOIN generate_series(0, 29) AS gs;
