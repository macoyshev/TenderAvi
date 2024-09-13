CREATE TABLE employee (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    username VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
);

CREATE TABLE organization (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type organization_type,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_responsible (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid (),
    organization_id UUID REFERENCES organization(id) ON DELETE CASCADE,
    user_id UUID REFERENCES employee(id) ON DELETE CASCADE
);

INSERT INTO employee (id, username, first_name, last_name)
VALUES 
('fd0d57a0-e74b-424e-a822-8ddf54c20b80', 'BrookeMcbride', 'Brooke', 'Mcbride'),
('7374388d-10a5-413a-8f2b-e96cef8f36f5', 'CarrieRamirez', 'Carrie', 'Ramirez'),
('403ac136-d3cb-4ed3-96e8-480387c74bcf', 'EricCabrera', 'Eric', 'Cabrera'),
('3d8d8286-f450-48a8-95f4-d41ff238aaeb', 'ElizabethStrickland', 'Elizabeth', 'Strickland'),
('a7779804-36de-4d11-b987-3ca6961a4059', 'MichaelAlexander', 'Michael', 'Alexander'),
('f00933a4-e12b-4f29-a14d-c9291eec7c86', 'RobertErickson', 'Robert', 'Erickson'),
('9f4c57b1-fd70-4512-b08b-877d23e11d35', 'RobertSimmons', 'Robert', 'Simmons'),
('b31847b2-9954-43b8-a69f-0fbc1d8a772c', 'KevinAshley', 'Kevin', 'Ashley'),
('1d011dfe-8b86-4f46-92d2-db50a151d4bd', 'JasonAnderson', 'Jason', 'Anderson'),
('de666809-e6f2-42db-bd98-d8450419c942', 'JamesSnyder', 'James', 'Snyder');

INSERT INTO organization (id, name, description, type)
VALUES 
('aff3999d-6c2c-4085-a63b-7c406df5d80b', 'Mention', 'Air move brother ground statio', 'LLC'),
('b322ef0e-86d0-4f1c-a6d3-6b62e55d02ca', 'Although', 'Anyone image herself garden cu', 'JSC'),
('6bbbbd31-e778-48b3-8adc-79790cba75e5', 'Surface', 'Business the need successful. ', 'LLC'),
('17394194-84ef-4c42-bf6b-b490a932b31e', 'Son', 'Sport protect movie event full', 'IE'),
('6539d3c1-1062-4714-a4e9-8267047a67b0', 'Form', 'Claim young nearly something s', 'LLC');

INSERT INTO organization_responsible (organization_id, user_id)
VALUES 
('aff3999d-6c2c-4085-a63b-7c406df5d80b', 'fd0d57a0-e74b-424e-a822-8ddf54c20b80'),
('aff3999d-6c2c-4085-a63b-7c406df5d80b', '7374388d-10a5-413a-8f2b-e96cef8f36f5'),
('b322ef0e-86d0-4f1c-a6d3-6b62e55d02ca', '403ac136-d3cb-4ed3-96e8-480387c74bcf'),
('b322ef0e-86d0-4f1c-a6d3-6b62e55d02ca', '3d8d8286-f450-48a8-95f4-d41ff238aaeb'),
('6bbbbd31-e778-48b3-8adc-79790cba75e5', 'a7779804-36de-4d11-b987-3ca6961a4059'),
('6bbbbd31-e778-48b3-8adc-79790cba75e5', 'f00933a4-e12b-4f29-a14d-c9291eec7c86'),
('17394194-84ef-4c42-bf6b-b490a932b31e', '9f4c57b1-fd70-4512-b08b-877d23e11d35'),
('17394194-84ef-4c42-bf6b-b490a932b31e', 'b31847b2-9954-43b8-a69f-0fbc1d8a772c'),
('6539d3c1-1062-4714-a4e9-8267047a67b0', '1d011dfe-8b86-4f46-92d2-db50a151d4bd'),
('6539d3c1-1062-4714-a4e9-8267047a67b0', 'de666809-e6f2-42db-bd98-d8450419c942');

