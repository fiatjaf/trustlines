### On a payment that involves A->B->C what happens if C withholds the hash from B, but later tries to claim its debt, showing the hashed contract and the preimage?

B is free to pay or to not pay, as always. These hashed contracts are not legally binding or anything. Remember this is a network of debts between trusted friends.

### On a payment that involves A->B->C->D what ensures D will not cheat and pass the preimage to B while not passing it to C?

Nothing, but if C doesn't get to know the preimage then he is free to ignore his payment to D in the actual world. Remember we're not transferring money here, just promises based on trust between friends.

### On a payment that involves X->Z->A->B->C what happens if B tries to pass the preimage back to A, but A is offline or never responds, even when B tries again later, and then A claims that it will not pay B because B was dishonest and didn't pass the preimage correctly.

A and B can disagree, but that disagreement would have happened in the real world anyway, the hash/preimage thing is just a computer trick to ensure everybody is talking about the same transaction. If A is dishonest, it wouldn't pay B anyway, and if B is dishonest it doesn't deserve to be paid anyway.

That leaves us with Z and its relation to A. If A was failing due to lack of connectivity or other honest reasons, B (if honest) could pass the preimage to Z and Z to X and so on, fulfilling the payment and making everybody happy if Z is honest and agrees to pay A instead of trying to cancel the hashed contract.

If Z is dishonest, A can lose, but that would have happened anyway, because A trusts in Z.
