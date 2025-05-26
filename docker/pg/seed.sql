DO $$
DECLARE
    i INTEGER;
    cities TEXT[] := ARRAY['Москва', 'Санкт-Петербург', 'Новосибирск', 'Екатеринбург', 'Казань'];
    first_names TEXT[] := ARRAY['Иван', 'Алексей', 'Дмитрий', 'Сергей', 'Андрей', 'Михаил', 'Артем', 'Максим', 'Александр', 'Евгений'];
    last_names TEXT[] := ARRAY['Иванов', 'Петров', 'Сидоров', 'Смирнов', 'Кузнецов', 'Попов', 'Васильев', 'Федоров', 'Морозов', 'Волков'];
    genders TEXT[] := ARRAY['male', 'female', 'other'];
    interests_list TEXT[] := ARRAY['спорт', 'музыка', 'кино', 'чтение', 'путешествия', 'программирование', 'фотография', 'кулинария'];
BEGIN
    FOR i IN 1..1000000 LOOP
        INSERT INTO users (
            first_name,
            last_name,
            email,
            birth_date,
            gender,
            interests,
            city,
            password_hash,
            created_at,
            updated_at
        ) VALUES (
            first_names[1 + floor(random() * array_length(first_names, 1))::int],
            last_names[1 + floor(random() * array_length(last_names, 1))::int],
            'user' || i || '@example.com',
            CURRENT_DATE - (365 * (18 + floor(random() * 50))::int)::integer,
            genders[1 + floor(random() * array_length(genders, 1))::int],
            (SELECT array_agg(interests_list[1 + floor(random() * array_length(interests_list, 1))::int])
             FROM generate_series(1, (2 + floor(random() * 5))::int)),
            cities[1 + floor(random() * array_length(cities, 1))::int],
            md5(random()::text),
            NOW() - (random() * 365)::integer * INTERVAL '1 day',
            NOW() - (random() * 365)::integer * INTERVAL '1 day'
        );

        IF i % 1000 = 0 THEN
            RAISE NOTICE 'Inserted % records', i;
        END IF;
    END LOOP;
END;
$$;