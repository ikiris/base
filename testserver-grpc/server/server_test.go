package server

func TestServer(t *testing.T) {
	ctx := context.Background()

	server := New("something")

	tests := map[string]struct{
		name string
		in
	}
}