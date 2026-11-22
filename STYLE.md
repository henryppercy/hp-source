# Comment Style Guide

1. Doc comments start with the name of the thing, read as a full sentence, and end with a full stop. This applies to every doc comment, exported or not.
2. Every exported identifier has a doc comment. Unexported identifiers get one only when non-obvious.
3. The first sentence says what the thing does or guarantees, in terms the signature doesn't already give. Any further text covers only what would surprise a caller ie concurrency safety, nil handling, wrapped errors, ownership of returned values.
4. Implementation reasoning belongs inline in the function body, never in the doc comment.

