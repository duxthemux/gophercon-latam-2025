-- +goose Up

create table kpis
(
    id    BIGSERIAL primary key,
    kpi   text,
    dt    TIMESTAMP,
    value REAL
);


INSERT INTO public.kpis (id, kpi, dt, value) VALUES (1, 'INCIDENTS', '2024-01-01 00:00:00.000000', 2);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (2, 'INCIDENTS', '2024-02-01 00:00:00.000000', 7);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (3, 'INCIDENTS', '2024-03-01 00:00:00.000000', 13);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (4, 'INCIDENTS', '2024-04-01 00:00:00.000000', 4);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (5, 'INCIDENTS', '2024-05-01 00:00:00.000000', 6);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (6, 'INCIDENTS', '2024-06-01 00:00:00.000000', 2);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (7, 'INCIDENTS', '2024-07-01 00:00:00.000000', 1);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (8, 'INCIDENTS', '2024-08-01 00:00:00.000000', 7);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (9, 'INCIDENTS', '2024-09-01 00:00:00.000000', 9);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (10, 'INCIDENTS', '2024-10-01 00:00:00.000000', 8);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (11, 'INCIDENTS', '2024-11-01 00:00:00.000000', 5);
INSERT INTO public.kpis (id, kpi, dt, value) VALUES (12, 'INCIDENTS', '2024-12-01 00:00:00.000000', 2);


-- +goose Down
drop table kpis