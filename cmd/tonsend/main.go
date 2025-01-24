package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func main() {
	app := &cli.App{
		Name:  "tonsend",
		Usage: "tonsend utilities in go",
		Commands: []*cli.Command{
			{
				Name:        "cnft-merkle",
				Description: "build merkle trees for compressed nft",
				ArgsUsage:   "[campaign name]",
				Action: func(cCtx *cli.Context) error {
					campaign := cCtx.Args().First()
					if campaign == "" {
						return errors.New("campaign missing")
					}

					return BuildCNFT(campaign, cCtx.Uint64("start"), cCtx.Uint64("end"), cCtx.Int("limit"))
				},
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "limit", Value: 200_000},
					&cli.Uint64Flag{Name: "start", Value: uint64(time.Now().Unix())},
					&cli.Uint64Flag{Name: "end", Value: uint64(time.Now().Add(365 * 24 * time.Hour).Unix())},
				},
			},
			{
				Name:        "cnft-merkle-proof",
				Description: "build merkle proof for compressed nft",
				ArgsUsage:   "[campaign name] [nft id]",
				Action: func(cCtx *cli.Context) error {
					campaign := cCtx.Args().First()
					if campaign == "" {
						return errors.New("campaign missing")
					}

					id := cCtx.Args().Get(1)
					if id == "" {
						return errors.New("id missing")
					}

					return BuildCNFTProof(campaign, id)
				},
			},
			{
				Name:      "crc-hex",
				ArgsUsage: "schema",
				Action: func(cCtx *cli.Context) error {
					schema := cCtx.Args().First()
					schema = strings.ReplaceAll(schema, "(", "")
					schema = strings.ReplaceAll(schema, ")", "")
					data := []byte(schema)
					crc := crc32.Checksum(data, crc32.MakeTable(crc32.IEEE))

					b_data := make([]byte, 4)
					binary.BigEndian.PutUint32(b_data, crc)
					res := hex.EncodeToString(b_data)
					log.Println(schema, fmt.Sprintf("0x%s", res))
					return nil
				},
			},
			{
				Name: "mnemonic",
				Action: func(cCtx *cli.Context) error {
					log.Println(strings.Join(wallet.NewSeed(), " "))
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}
