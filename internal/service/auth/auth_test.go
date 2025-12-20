package auth

import (
	"github.com/gobugger/gomarket/internal/repo"
	"github.com/gobugger/gomarket/internal/testutil"
	"github.com/gobugger/gomarket/pkg/pgp"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAuth(t *testing.T) {
	infra := testutil.NewInfra(t.Context())

	creds := []struct {
		username     string
		password     string
		pgpkey       string
		registerFail bool
	}{
		{
			username: "pirkka",
			password: "pirkka123",
		},
		{
			username: "pirkko",
			password: "pirkko123",
			pgpkey:   testutil.PgpKey,
		},
		{
			username:     "",
			password:     "pirkko123",
			pgpkey:       "invalid pgp key",
			registerFail: true,
		},
		{
			username:     "timppa",
			password:     "",
			registerFail: true,
		},
		{
			username:     "pirkko",
			password:     "pirkko123",
			pgpkey:       "invalid pgp key",
			registerFail: true,
		},
	}

	ctx := t.Context()
	q := repo.New(infra.Db)

	for _, cred := range creds {
		u, err := Register(ctx, q, RegisterParams{
			Username: cred.username,
			Password: cred.password,
			PgpKey:   cred.pgpkey,
		})

		if cred.registerFail {
			require.Error(t, err, "registeration should have failed for %v", cred)
			continue
		}

		_, err = Register(ctx, q, RegisterParams{
			Username: cred.username,
			Password: "SomePrettyGoodPassword123!!??",
			PgpKey:   cred.pgpkey,
		})
		require.ErrorIs(t, err, ErrUsernameAlreadyRegistered)

		_, err = Authenticate(ctx, q, AuthenticateParams{
			Username: cred.username + "a",
			Password: cred.password,
		})
		require.ErrorIs(t, err, ErrInvalidCredentials)

		_, err = Authenticate(ctx, q, AuthenticateParams{
			Username: cred.username,
			Password: cred.password + "a",
		})
		require.ErrorIs(t, err, ErrInvalidCredentials)

		u2, err := Authenticate(ctx, q, AuthenticateParams{
			Username: cred.username,
			Password: cred.password,
		})
		require.NoError(t, err)

		require.Equal(t, u.ID, u2.ID)

		_, err = q.CreateBan(ctx, u.ID)
		require.NoError(t, err)

		_, err = Authenticate(ctx, q, AuthenticateParams{
			Username: cred.username,
			Password: cred.password,
		})
		require.ErrorIs(t, err, ErrAccountIsBanned)
	}
}

func Test2FA(t *testing.T) {
	tests := []struct {
		key string
		ok  bool
	}{
		{
			key: testutil.PgpKey,
			ok:  true,
		},
		{
			key: "badkey",
			ok:  false,
		},
		{
			key: "",
			ok:  false,
		},
	}

	for _, test := range tests {
		c, err := Generate2FAChallenge(test.key)
		if !test.ok {
			if err == nil {
				t.Fatal()
			} else {
				continue
			}

		}

		if err != nil {
			t.Fatal()
		}

		if c == nil || c.Token == "" {
			t.Fatal()
		}
	}
}

func TestChangePassword(t *testing.T) {
	infra := testutil.NewInfra(t.Context())

	ctx := t.Context()
	q := repo.New(infra.Db)

	u, err := Register(ctx, q, RegisterParams{
		Username: "pirkka",
		Password: "pirkka123",
		PgpKey:   "",
	})
	require.NoError(t, err)

	_, err = ChangePassword(ctx, q, ChangePasswordParams{
		Username:    "randomname",
		OldPassword: "pirkka123",
		NewPassword: "pirkka1234",
	})
	require.ErrorIs(t, err, ErrInvalidCredentials)

	_, err = ChangePassword(ctx, q, ChangePasswordParams{
		Username:    u.Username,
		OldPassword: "pirkka321",
		NewPassword: "pirkka1234",
	})
	require.ErrorIs(t, err, ErrInvalidPassword)

	_, err = ChangePassword(ctx, q, ChangePasswordParams{
		Username:    u.Username,
		OldPassword: "pirkka123",
		NewPassword: "pirkka123",
	})
	require.ErrorIs(t, err, ErrInvalidNewPassword)

	u2, err := ChangePassword(ctx, q, ChangePasswordParams{
		Username:    u.Username,
		OldPassword: "pirkka123",
		NewPassword: "pirkka321",
	})
	require.NoError(t, err)
	require.Equal(t, u.ID, u2.ID)

	u2, err = Authenticate(ctx, q, AuthenticateParams{
		Username: u.Username,
		Password: "pirkka321",
	})
	require.NoError(t, err)
	require.Equal(t, u.ID, u2.ID)
}

func TestSetPGPKey(t *testing.T) {
	infra := testutil.NewInfra(t.Context())

	ctx := t.Context()
	q := repo.New(infra.Db)

	pirkka, err := Register(ctx, q, RegisterParams{
		Username: "pirkka",
		Password: "pirkka123",
		PgpKey:   "",
	})
	require.NoError(t, err)

	_, err = SetPGPKey(ctx, q, SetPGPKeyParams{
		UserID: pirkka.ID,
		PgpKey: "",
	})
	require.ErrorIs(t, err, ErrInvalidPGPKey)

	_, err = SetPGPKey(ctx, q, SetPGPKeyParams{
		UserID: pirkka.ID,
		PgpKey: "badpgpkey",
	})
	require.ErrorIs(t, err, ErrInvalidPGPKey)

	u, err := SetPGPKey(ctx, q, SetPGPKeyParams{
		UserID: pirkka.ID,
		PgpKey: testutil.PgpKey,
	})
	require.NoError(t, err)
	require.Equal(t, pirkka.ID, u.ID)
	require.True(t, pgp.PublicKeyIsValid(u.PgpKey))
	require.Equal(t, testutil.PgpKey, u.PgpKey)
}
