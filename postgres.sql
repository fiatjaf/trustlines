CREATE TABLE users (
  id text PRIMARY KEY,
  email text UNIQUE NOT NULL,
  passwordhash text UNIQUE NOT NULL
);

CREATE TABLE money_transfers (
  -- this may mean that
  -- a) <target> has borrowed <amount> of <currency> from <source> (debts)
  -- b) <source> has paid <amount> of <currency> for a good for <target> (debts)
  -- c) <source> has given <target> money in repayment of a previous debt (settlements)
  -- d) <source> is charging interest on <target> (interest_charges)
  -- e) <target> has proposed a hashed contract on <source> (hashed_contracts)

  id text PRIMARY KEY, -- unique identifier
  target text NOT NULL, -- username@server.name
  source text NOT NULL, -- username@server.name
  t timestamp NOT NULL DEFAULT now(), -- server date
  amount int NOT NULL, -- in the smallest enumerable, 5 USD is actually 500 here
  currency text NOT NULL, -- 
  description text NOT NULL,
  signature text NOT NULL, -- of source:target:t:amount:currency + whatever

  CHECK (source != target)
);

CREATE TABLE debts (
  typ text NOT NULL DEFAULT 'Debt',
  actual_date date NOT NULL
) INHERITS (money_transfers);

CREATE TABLE settlements (
  typ text NOT NULL DEFAULT 'Settlement',
  actual_date date NOT NULL
) INHERITS (money_transfers);

CREATE TABLE interest_charges (
  typ text NOT NULL DEFAULT 'Interest',
  actual_date date NOT NULL,
  over_amount int
) INHERITS (money_transfers);

CREATE TABLE hashed_contracts (
  typ text NOT NULL DEFAULT 'HashedContract',
  timeout interval NOT NULL DEFAULT '1 DAY', -- after this time the presenting the
                                             -- preimage doesn't have to work anymore
                                             -- but the person who's on the debtor side
                                             -- has to manually cancel the contract.
  cancelled boolean NOT NULL DEFAULT false,
  hash text NOT NULL, -- the creditor possessing this makes the debt a reality
  preimage TEXT NOT NULL,
  next TEXT, -- the uri of the next actor in the chain of hashed contracts
             -- (if source is B and target is A in A->B->C, then this will be C)

  CHECK (cancelled = true AND t + timeout < now())
) INHERITS (money_transfers);

