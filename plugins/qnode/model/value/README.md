### UTXO model

UTXO transfer consists of:

- N **inputs**
- M **outputs**
- **signatures**

UTXO transfer is part of the value transaction. It can be located by transaction ID.

#### Inputs
`Inputs` is an array: `input[0]`, `input[1]`, ..., `input[N-1]`

Individual input within the UTXO transfer is uniquely identified by its **index**.

Sequence of inputs within the transfer is fixed and deterministic. 

Each `input[i]` is an **output reference**, the reference to some output in another transfer (see below).

#### Outputs
`outputs` is an array: `output[0]`, `output[1]`, ..., `output[M-1]`

Each `output[i]` is `address` and `balance`

`output` ::= (`address`, `balance`)

Individual output within the UTXO transfer is uniquely identified by its **index** in the array.
Sequence of outputs within the transfer is fixed and deterministic. 

Outside of the UTXO transaction is uniquely identified by `output reference`: 
`transaction id` and `output index` 
within that transaction. 

`output ref` ::= (`txid`, `output index`)

An input of the UTXO transfer always references corresponding output in another UTXO transfer by its output reference. 


`transaction id` is 32 bytes, index is 2 bytes (2^32 maximum number of outputs with the transfer)

So reference to output (= input) will take 34 bytes.

#### Database of outputs

The node will have to maintain index of all outputs. 
There two main functions to access it:

- `getOutputByOutputReference(txid, outputIdx)` it returns (`address`, `balance`). It is needed for example for signing   
- `getUnspentOutputsForAddress(address)` it returns all unspent outputs. It is needed for the wallet and similar.

#### Signing of the UTXO transfer

1. given array of inputs, collect addresses of referenced outputs
2. sort inputs by referenced addresses
3. group inputs by addresses
4. produce signature for each address

Result is a deterministic order of signatures

#### Validation of the transfer
For an UTXO transfer to be correct the sum of input balances must be equal to the sum of output balances.
Balances of outputs are in the transaction.

Balances of inputs must be collected via output database using outputs references in the inputs. 

#### Determinism

Result is deterministic order of inputs, outputs and signatures within the transaction. It also remains deterministically ordered when mashalling/unmarshalling

No need for Go maps and more complicated structures to solve a problem with non-deterministic order of iteration along Go maps.    

Almost 2 times less bytes needed for inputs -> significantly smaller value transactions

