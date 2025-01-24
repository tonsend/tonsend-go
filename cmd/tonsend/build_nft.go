package main

import (
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton/nft"
	"github.com/xssnick/tonutils-go/tvm/cell"
	"golang.org/x/sync/errgroup"
)

type Handle interface {
	RowToNFT(record []string) (*big.Int, *address.Address, *cell.Cell, error)
	IDBigInt(str string) *big.Int
}

type DefaultHandler struct {
	Name        string
	Image       string
	Description string
}

var handlers = map[string]Handle{
	"default": DefaultHandler{},
}

func BuildCNFT(campaign string, start, end uint64, limit int) error {
	handler := handlers["default"]
	fileInput := filepath.Join(fmt.Sprintf("./input/%s.csv", campaign))
	dirOutput := filepath.Join(fmt.Sprintf("./output/%s", campaign))

	if _, err := os.Stat(dirOutput); os.IsNotExist(err) {
		if err := os.Mkdir(dirOutput, 0o700); err != nil {
			return err
		}
	}

	log.Println(campaign, "start:", start, ", end:", end, limit, fileInput, dirOutput)

	f, err := os.Open(fileInput)
	if err != nil {
		return err
	}
	defer f.Close()

	// TODO: stream
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	i := 0
	remainder := 0
	reader := csv.NewReader(bytes.NewReader(data))
	dict := cell.NewDict(256)
	keys := map[string]int{} // more compact than bool when marshalling

	for {
		remainder = i % limit

		if remainder == limit-1 {
			// instead of regular flush, we clone and flush the batch inside goroutine
			// this will result in a higher memory consumption but better performance
			dictClone := dict.Copy()
			keysClone := map[string]int{}
			for k, v := range keys {
				keysClone[k] = v
			}

			// nolint:errcheck
			go writeDict(dirOutput, dictClone, keysClone)

			dict = cell.NewDict(256)
			keys = map[string]int{}
		}

		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		id, addr, metadata, err := handler.RowToNFT(record)
		if err != nil {
			return err
		}

		keys[id.String()] = 1 // true
		err = dict.SetIntKey(id, cell.BeginCell().MustStoreAddr(addr).MustStoreRef(metadata).MustStoreUInt(start, 48).MustStoreUInt(end, 48).EndCell())
		if err != nil {
			return err
		}

		i++
	}

	if limit-remainder > 1 {
		_, _, err := writeDict(dirOutput, dict, keys)
		return err
	}

	return nil
}

func BuildCNFTProof(campaign string, id string) error {
	dirOutput := filepath.Join(fmt.Sprintf("./output/%s", campaign))
	re := regexp.MustCompile(`(.*?)\/([a-z0-9]+)\.json`)

	handler := handlers["default"]
	files, err := filepath.Glob(dirOutput + "/*.json")
	if err != nil {
		return err
	}

	k := handler.IDBigInt(id)

	g := new(errgroup.Group)

	found := []string{}

	for _, file := range files {
		g.Go(func() error {
			b, err := os.ReadFile(file)
			if err != nil {
				return err
			}

			m := map[string]int{}
			err = json.Unmarshal(b, &m)
			if err != nil {
				return err
			}

			if _, ok := m[k.String()]; ok {
				found = append(found, re.ReplaceAllString(file, "$2"))
			}

			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return err
	}

	if len(found) == 0 {
		return errors.New("id not found")
	}

	log.Println("found", len(found), found)

	for _, root := range found {
		g.Go(func() error {
			b, err := os.ReadFile(filepath.Join(fmt.Sprintf("./output/%s/%s.boc", campaign, root)))
			if err != nil {
				return err
			}

			c, err := cell.FromBOC(b)
			if err != nil {
				return err
			}

			dict, err := c.BeginParse().ToDict(256)
			if err != nil {
				return err
			}
			kProof := cell.BeginCell().MustStoreBigInt(k, 256).EndCell()
			sk := cell.CreateProofSkeleton()
			_, skk, err := dict.LoadValueWithProof(kProof, sk)
			if err != nil {
				return err
			}

			skk.SetRecursive()
			proof, err := dict.AsCell().CreateProof(sk)
			if err != nil {
				return err
			}

			log.Println(root, hex.EncodeToString(proof.ToBOC()))

			return nil
		})
	}

	if err = g.Wait(); err != nil {
		return err
	}

	return nil
}

func writeDict(dirOutput string, dict *cell.Dictionary, keys map[string]int) (string, string, error) {
	root := hex.EncodeToString(dict.AsCell().Hash(0))
	log.Println(root)
	fileBOC := filepath.Join(dirOutput, fmt.Sprintf("%s.boc", root))
	fileJSON := filepath.Join(dirOutput, fmt.Sprintf("%s.json", root))

	g := new(errgroup.Group)

	g.Go(func() error {
		return os.WriteFile(fileBOC, dict.AsCell().ToBOC(), 0o644)
	})

	g.Go(func() error {
		b, err := json.Marshal(keys)
		if err != nil {
			return err
		}

		return os.WriteFile(fileJSON, b, 0o644)
	})

	return fileBOC, fileJSON, g.Wait()
}

func (v DefaultHandler) IDBigInt(str string) *big.Int {
	id := new(big.Int)
	id.SetString(strings.Replace(str, "-", "", 4), 16)
	return id
}

func (v DefaultHandler) RowToNFT(record []string) (*big.Int, *address.Address, *cell.Cell, error) {
	if len(record) < 4 {
		return nil, nil, nil, fmt.Errorf("invalid record: %v", record)
	}

	addr := address.MustParseAddr(record[0])
	id := v.IDBigInt(record[1])

	name := record[2]
	image := record[3]

	content := &nft.ContentOnchain{
		Name:  name,
		Image: image,
	}

	if len(record) == 5 {
		content.SetAttribute("description", record[4])
	}

	metadata, err := content.ContentCell()
	if err != nil {
		return nil, nil, nil, err
	}

	return id, addr, metadata, err
}
