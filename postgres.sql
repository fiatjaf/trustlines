CREATE TABLE users (
  id text PRIMARY KEY,
  email text UNIQUE NOT NULL,
  passwordhash text UNIQUE NOT NULL
);

drop table money_transfers;

CREATE TABLE transfers (
  -- this may mean that
  -- a) <target> has borrowed <amount> of <currency> from <source> (debts)
  -- b) <source> has paid <amount> of <currency> for a good for <target> (debts)
  -- c) <source> has given <target> money in repayment of a previous debt (settlements)
  -- d) <source> is charging interest on <target> (interest charges)
  -- e) <target> will owe <sourc>e if <source> presents some condition (conditional debt)

  id text PRIMARY KEY, -- unique identifier
  target text NOT NULL, -- username@server.name
  source text NOT NULL, -- username@server.name
  t timestamp NOT NULL DEFAULT now(), -- server date
  amount int NOT NULL, -- in the smallest enumerable, 5 USD is actually 500 here
  currency text NOT NULL, -- 
  description text NOT NULL,
  signature text NOT NULL, -- of source:target:t:amount:currency + whatever
  actual_date datetime,
  condition jsonb,

  CHECK (source != target)
);

CREATE TABLE trustlines (
  truster text NOT NULL, -- someone@some.where
  trusted text NOT NULL, -- someone@some.where
  currency text NOT NULL,
  amount text NOT NULL,

  UNIQUE(truster, trusted, currency)
);
