// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package ach

// BatchPOP holds the BatchHeader and BatchControl and all EntryDetail for POP Entries.
//
// Point-of-Purchase. A check presented in-person to a merchant for purchase is presented
// as an ACH entry instead of a physical check.
//
// This ACH debit application is used by originators as a method of payment for the
// in-person purchase of goods or services by consumers. These Single Entry debit
// entries are initiated by the originator based on a written authorization and
// account information drawn from the source document (a check) obtained from the
// consumer at the point-of-purchase. The source document, which is voided by the
// merchant and returned to the consumer at the point-of-purchase, is used to
// collect the consumer’s routing number, account number and check serial number that
// will be used to generate the debit entry to the consumer’s account.
//
// The difference between POP and ARC is that ARC can result from a check mailed in whereas POP is in-person.
type BatchPOP struct {
	Batch
}

// NewBatchPOP returns a *BatchPOP
func NewBatchPOP(bh *BatchHeader) *BatchPOP {
	batch := new(BatchPOP)
	batch.SetControl(NewBatchControl())
	batch.SetHeader(bh)
	return batch
}

// Validate checks properties of the ACH batch to ensure they match NACHA guidelines.
// This includes computing checksums, totals, and sequence orderings.
//
// Validate will never modify the batch.
func (batch *BatchPOP) Validate() error {
	// basic verification of the batch before we validate specific rules.
	if err := batch.verify(); err != nil {
		return err
	}

	// Add configuration and type specific validation for this type.
	if batch.Header.StandardEntryClassCode != POP {
		return batch.Error("StandardEntryClassCode", ErrBatchSECType, POP)
	}

	// POP detail entries can only be a debit, ServiceClassCode must allow debits
	switch batch.Header.ServiceClassCode {
	case MixedDebitsAndCredits, CreditsOnly:
		return batch.Error("ServiceClassCode", ErrBatchServiceClassCode, batch.Header.ServiceClassCode)
	}

	for _, entry := range batch.Entries {
		// POP detail entries must be a debit
		if entry.CreditOrDebit() != "D" {
			return batch.Error("TransactionCode", ErrBatchDebitOnly, entry.TransactionCode)
		}
		// Amount must be 25,000 or less
		if entry.Amount > 2500000 {
			return batch.Error("Amount", NewErrBatchAmount(entry.Amount, 2500000))
		}
		// CheckSerialNumber, Terminal City, Terminal State underlying IdentificationNumber, must be defined
		if entry.IdentificationNumber == "" {
			return batch.Error("CheckSerialNumber", ErrBatchCheckSerialNumber)
		}
		// Verify the TransactionCode is valid for a ServiceClassCode
		if err := batch.ValidTranCodeForServiceClassCode(entry); err != nil {
			return err
		}
		// Verify Addenda* FieldInclusion based on entry.Category and batchHeader.StandardEntryClassCode
		if err := batch.addendaFieldInclusion(entry); err != nil {
			return err
		}
	}
	return nil
}

// Create will tabulate and assemble an ACH batch into a valid state. This includes
// setting any posting dates, sequence numbers, counts, and sums.
//
// Create implementations are free to modify computable fields in a file and should
// call the Batch's Validate() function at the end of their execution.
func (batch *BatchPOP) Create() error {
	// generates sequence numbers and batch control
	if err := batch.build(); err != nil {
		return err
	}
	// Additional steps specific to batch type
	// ...

	return batch.Validate()
}
