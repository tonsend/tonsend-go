# tonsend-go

TonSend utilities in go

### CNFT Merkle Tree

1. Create a csv file inside input directory with data given in format of `owner,id,name,image,description (optional)`.  
The ID could be any hex-format uint256 number.

`input/foobar.csv`
```
EQBiA46W-PQaaZZNFIDglnVknV9CR6J5hs81bSv70FwfNTrD,f477a680-c1a1-4758-ae87-4a0d04e946f0,Item #1,ipfs://cid/f477a680-c1a1-4758-ae87-4a0d04e946f0.png
EQBc3CG3NOeF3wwkBM8zjXrsWUhjLuN45LobSkZHXCR0jhvg,bc382319-a5b0-4b5e-a608-a498aec496b5,Item #2,ipfs://cid/bc382319-a5b0-4b5e-a608-a498aec496b5.png
EQBmI8Qf6dyMpQqOTAsKD99zelqzSpJNlZ7V5RDwDFFD1oJ6,88fa1dd5-eb05-4f03-8586-559aa934f171,Item #3,ipfs://cid/88fa1dd5-eb05-4f03-8586-559aa934f171.png
EQAsqufgnT0d2zGwXHSwsLTgnO8wDhvAY4OhIUEj02QXZAbf,a84822bf-e0bf-4a37-a120-1d531ff346c7,Item #4,ipfs://cid/a84822bf-e0bf-4a37-a120-1d531ff346c7.png
EQBtcUiel-vjJMdmiYhye4VbRlPYLoiPcbYfQ-w8k6y3bbHB,60b23a92-2c04-49a8-b3e0-42d03b67ffc5,Item #5,ipfs://cid/60b23a92-2c04-49a8-b3e0-42d03b67ffc5.png
EQDVvqQi3QI14FIgP5tjULpZMpQBtsPcHVOqTeffKI-TKpob,b4956b28-65e0-4f51-95f9-e7da891a1cd0,Item #6,ipfs://cid/b4956b28-65e0-4f51-95f9-e7da891a1cd0.png
EQBFhmn9OYHoPkHsfJp5eiEM_yqIv_VBVj_jalFWAOHnxxFG,445274a5-8fe6-403c-b2fb-d47db108cd74,Item #7,ipfs://cid/445274a5-8fe6-403c-b2fb-d47db108cd74.png
EQDsdyN5ziWcKI9EYMQfpFTvVyacA-QChnsOY66RiT71Z1TS,35514abf-585b-4582-841a-de0f8c910e12,Item #8,ipfs://cid/35514abf-585b-4582-841a-de0f8c910e12.png
EQA0RLIJzwY2Jj99XobCWG5HqKF8KlSi7MR6T-C1N8sOfMPs,c9080c60-4138-4e0c-8f47-78d03acb5954,Item #9,ipfs://cid/c9080c60-4138-4e0c-8f47-78d03acb5954.png
EQC5H95YET5c6Jpu9ivo9Ex2qX36CnLQEaDtmuQioVfkLBij,8422bcf2-655a-4393-86aa-68c6ac3a1145,Item #10,ipfs://cid/8422bcf2-655a-4393-86aa-68c6ac3a1145.png
```

2. Build the tree
```bash
go run cmd/tonsend/*.go cnft-merkle foobar --limit 200000 --start 1737676800 --end 1895443200
```

It should output
```
2025/01/24 17:58:24 foobar start: 1737716304 , end: 1769252304 200000 input/foobar.csv output/foobar
2025/01/24 17:58:24 b6c76ba120be945fd38f1bd22f3bac73d795cbff01426eafa0eb6ade38d185c3
```

The latter line is the merkle root for the chunk.

`limit`: Maximum number of items per chunk.  
`start`: Claimable after start. In second. Set it to `0` to disable start time checking.  
`end`: Un-claimable after end. In second. Set it to `0` to disable end time checking.  

3. Generate proof for an item given its ID.
```bash
go run cmd/tonsend/*.go cnft-merkle-proof foobar 88fa1dd5-eb05-4f03-8586-559aa934f171
```

This command will look for the item ID within the output directory of campaign foobar. Once it locates the chunk of the item, it will build the merkle proof and dump the BOC of the proof.