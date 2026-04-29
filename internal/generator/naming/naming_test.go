package naming

import "testing"

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"public_forum", "PublicForum"},
		{"user", "User"},
		{"bookingsvc/booking", "BookingsvcBooking"},
		{"bookingsvc/user", "BookingsvcUser"},
		{"menusvc/company", "MenusvcCompany"},
		{"simple", "Simple"},
		{"already_pascal", "AlreadyPascal"},
		{"a_b_c", "ABC"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToPascalCase(tt.input)
			if got != tt.want {
				t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"public_forum", "publicForum"},
		{"user", "user"},
		{"bookingsvc/booking", "bookingsvcBooking"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToCamelCase(tt.input)
			if got != tt.want {
				t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"bookingsvc/booking", "bookingsvc_booking"},
		{"public_forum", "public_forum"},
		{"User", "user"},
		{"menusvc/company", "menusvc_company"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ToSnakeCase(tt.input)
			if got != tt.want {
				t.Errorf("ToSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTypeConstName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"bookingsvc/booking", "TypeBookingsvcBooking"},
		{"public_forum", "TypePublicForum"},
		{"user", "TypeUser"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TypeConstName(tt.input)
			if got != tt.want {
				t.Errorf("TypeConstName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRelationConstName(t *testing.T) {
	tests := []struct {
		defName string
		relName string
		want    string
	}{
		{"public_forum", "owner", "PublicForumRelationOwner"},
		{"bookingsvc/booking", "creator", "BookingsvcBookingRelationCreator"},
	}
	for _, tt := range tests {
		t.Run(tt.defName+"_"+tt.relName, func(t *testing.T) {
			got := RelationConstName(tt.defName, tt.relName)
			if got != tt.want {
				t.Errorf("RelationConstName(%q, %q) = %q, want %q", tt.defName, tt.relName, got, tt.want)
			}
		})
	}
}

func TestPermissionConstName(t *testing.T) {
	tests := []struct {
		defName  string
		permName string
		want     string
	}{
		{"public_forum", "view", "PublicForumPermissionView"},
		{"bookingsvc/booking", "write", "BookingsvcBookingPermissionWrite"},
	}
	for _, tt := range tests {
		t.Run(tt.defName+"_"+tt.permName, func(t *testing.T) {
			got := PermissionConstName(tt.defName, tt.permName)
			if got != tt.want {
				t.Errorf("PermissionConstName(%q, %q) = %q, want %q", tt.defName, tt.permName, got, tt.want)
			}
		})
	}
}

func TestRelationObjectsStructName(t *testing.T) {
	tests := []struct {
		defName string
		relName string
		want    string
	}{
		{"public_forum", "owner", "PublicForumOwnerObjects"},
		{"bookingsvc/booking", "creator", "BookingsvcBookingCreatorObjects"},
	}
	for _, tt := range tests {
		t.Run(tt.defName+"_"+tt.relName, func(t *testing.T) {
			got := RelationObjectsStructName(tt.defName, tt.relName)
			if got != tt.want {
				t.Errorf("RelationObjectsStructName(%q, %q) = %q, want %q", tt.defName, tt.relName, got, tt.want)
			}
		})
	}
}

func TestCheckInputStructName(t *testing.T) {
	tests := []struct {
		defName  string
		permName string
		want     string
	}{
		{"public_forum", "view", "CheckPublicForumViewInputs"},
		{"bookingsvc/booking", "write", "CheckBookingsvcBookingWriteInputs"},
	}
	for _, tt := range tests {
		t.Run(tt.defName+"_"+tt.permName, func(t *testing.T) {
			got := CheckInputStructName(tt.defName, tt.permName)
			if got != tt.want {
				t.Errorf("CheckInputStructName(%q, %q) = %q, want %q", tt.defName, tt.permName, got, tt.want)
			}
		})
	}
}

func TestTypeStructName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"public_forum", "PublicForum"},
		{"bookingsvc/booking", "BookingsvcBooking"},
		{"user", "User"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TypeStructName(tt.input)
			if got != tt.want {
				t.Errorf("TypeStructName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestReceiverName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"PublicForum", "pf"},
		{"User", "u"},
		{"BookingsvcBooking", "bb"},
		{"MenusvcCompany", "mc"},
		{"ABCDType", "abc"}, // 5 uppercase → "abcdt", truncated to 3
		{"", "x"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ReceiverName(tt.input)
			if got != tt.want {
				t.Errorf("ReceiverName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
