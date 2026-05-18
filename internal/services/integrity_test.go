package services

import "testing"

func TestValidateJournalLinesRequiresBalancedPosting(t *testing.T) {
	err := ValidateJournalLines([]JournalLineInput{
		{AccountCode: "1000", Debit: 100},
		{AccountCode: "4000", Credit: 90},
	})
	if err == nil {
		t.Fatal("expected unbalanced journal to fail")
	}
}

func TestValidateJournalLinesRejectsInvalidLine(t *testing.T) {
	err := ValidateJournalLines([]JournalLineInput{
		{AccountCode: "1000", Debit: 100, Credit: 100},
		{AccountCode: "4000", Credit: 100},
	})
	if err == nil {
		t.Fatal("expected debit and credit on same line to fail")
	}
}

func TestValidatePasswordPolicy(t *testing.T) {
	if err := ValidatePasswordPolicy("short"); err == nil {
		t.Fatal("expected weak password to fail")
	}
	if err := ValidatePasswordPolicy("StrongPass1!"); err != nil {
		t.Fatalf("expected strong password to pass: %v", err)
	}
}
