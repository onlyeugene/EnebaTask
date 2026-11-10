❯ <ide_opened_file>The user opened the file /Users/onlyeugene/Desktop/Eneba
Marketplace/server/internal/products/repository_mongo.go in the IDE. This may
or may not be related to the current task.</ide_opened_file>

❯ lets move to phase 6 now

so read claude.md, plan.md and api.md first. then start phase 6 from plan.md.
stop when it's done and tell me how to check it, including the postman
requests.

⏺ I'll start by reading the project docs.

Read 3 files (ctrl+o to expand)

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && ls
internal/auth internal/platform/_ cmd/api && wc -l internal/auth/_.go
cmd/api/\*.go internal/platfo…)
⎿  (No output)

Read 12 files (ctrl+o to expand)

⏺ Phase 6 covers refresh rotation, reuse detection, logout, TTL indexes and
tests, and the TTL indexes already exist in indexes.go. Next I'll check how
auth is wired and which Go version is in use.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && head -5 go.mod
&& go version && grep -n "auth\." cmd/api/_.go && ls ../ | head -30 && ls
../postman_…)
⎿  Error: Exit code 1
module eugene

     go 1.25.4

     require (
     go version go1.25.4 darwin/arm64
     cmd/api/routes.go:23:   auth     *auth.Handler
     cmd/api/routes.go:51:           api.Mount("/auth", h.auth.Routes())
     cmd/api/main.go:50:     authSvc, err :=
     auth.NewService(auth.NewMongoUsers(database),
     auth.NewMongoRefresh(database), auth.NewMongoRevoked(database),
     … +12 lines (ctrl+o to expand)

⏺ Wiring in main.go needs no change. Next I'll add the repository methods.

⏺ Update(internal/auth/repository.go)
⎿  Added 3 lines, removed 2 lines
8 )
9  
 10 var (
11 - ErrUserNotFound = errors.New("user not found")  
 12 - ErrDuplicate = errors.New("email already in use")  
 11 + ErrUserNotFound = errors.New("user not found")  
 12 + ErrDuplicate = errors.New("email already in use")  
 13 + ErrRefreshNotFound = errors.New("refresh token not found")  
 14 )
15  
 16 type UserRepo interface {

⏺ Update(internal/auth/repository.go)
⎿  Added 12 lines
23  
 24 type RefreshRepo interface {
25 Insert(ctx context.Context, t RefreshToken) error
26 + // Use marks the token used and points it at its successor, i
+n one atomic step, but only if it  
 27 + // is unused and not expired. Two calls with the same token c
+an't both succeed: the loser gets  
 28 + // ErrRefreshNotFound.  
 29 + Use(ctx context.Context, hash string, now time.Time, replaced
+By bson.ObjectID) (RefreshToken, error)  
 30 + // ByHash finds a token whatever its state; ErrRefreshNotFoun
+d if there is none.  
 31 + ByHash(ctx context.Context, hash string) (RefreshToken, error
+)  
 32 + // DeleteUser deletes every refresh token of a user (reuse de
+tected: log out all sessions).  
 33 + DeleteUser(ctx context.Context, userID bson.ObjectID) error  
 34 + // DeleteFamily deletes one login session's chain, only if it + belongs to userID.  
 35 + DeleteFamily(ctx context.Context, userID, familyID bson.Objec
+tID) error  
 36 }
37  
 38 type RevokedRepo interface {
39 // IsRevoked reports whether an access token's jti was revoke
d by logout.
40 IsRevoked(ctx context.Context, jti string) (bool, error)
41 + // Revoke stores the jti until expiresAt (the token's own exp
+). Revoking twice is not an error.  
 42 + Revoke(ctx context.Context, jti string, userID bson.ObjectID, + expiresAt time.Time) error  
 43 }

⏺ Update(internal/auth/repository.go)
⎿  Added 1 line
3 import (
4 "context"
5 "errors"
6 + "time"  
 7  
 8 "go.mongodb.org/mongo-driver/v2/bson"
9 )

⏺ Now the Mongo side.

⏺ Update(internal/auth/repository*mongo.go)
⎿  Added 43 lines
61 return nil
62 }
63  
 64 +// Use is one FindOneAndUpdate, so the check (unused, not expi
+red) and the write (mark used) can't  
 65 +// be split by another request with the same token. expiresAt
+is checked here as well because the  
 66 +// TTL monitor only runs about once a minute.  
 67 +func (r *MongoRefresh) Use(ctx context.Context, hash string, n
+ow time.Time, replacedBy bson.ObjectID) (RefreshToken, error)
+{  
 68 + filter := bson.M{"tokenHash": hash, "usedAt": nil, "expiresA
+t": bson.M{"$gt": now}}                                       
       69 +  update := bson.M{"$set": bson.M{"usedAt": now, "replacedBy": + replacedBy}}  
 70 + var t RefreshToken  
 71 + err := r.tokens.FindOneAndUpdate(ctx, filter, update).Decode
+(&t)  
 72 + if errors.Is(err, mongo.ErrNoDocuments) {  
 73 + return RefreshToken{}, ErrRefreshNotFound  
 74 + }  
 75 + if err != nil {  
 76 + return RefreshToken{}, fmt.Errorf("use refresh token: %w", + err)  
 77 + }  
 78 + return t, nil  
 79 +}  
 80 +  
 81 +func (r *MongoRefresh) ByHash(ctx context.Context, hash string
+) (RefreshToken, error) {  
 82 + var t RefreshToken  
 83 + err := r.tokens.FindOne(ctx, bson.M{"tokenHash": hash}).Deco
+de(&t)  
 84 + if errors.Is(err, mongo.ErrNoDocuments) {  
 85 + return RefreshToken{}, ErrRefreshNotFound  
 86 + }  
 87 + if err != nil {  
 88 + return RefreshToken{}, fmt.Errorf("find refresh token: %w"
+, err)  
 89 + }  
 90 + return t, nil  
 91 +}  
 92 +  
 93 +func (r \*MongoRefresh) DeleteUser(ctx context.Context, userID
+bson.ObjectID) error {  
 94 + if *, err := r.tokens.DeleteMany(ctx, bson.M{"userId": userI
+D}); err != nil {  
 95 + return fmt.Errorf("delete user refresh tokens: %w", err)  
 96 + }  
 97 + return nil  
 98 +}  
 99 +  
 100 +func (r *MongoRefresh) DeleteFamily(ctx context.Context, userI
+D, familyID bson.ObjectID) error {  
 101 + if \_, err := r.tokens.DeleteMany(ctx, bson.M{"userId": userI
+D, "familyId": familyID}); err != nil {  
 102 + return fmt.Errorf("delete refresh family: %w", err)  
 103 + }  
 104 + return nil  
 105 +}  
 106 +  
 107 type MongoRevoked struct{ revoked *mongo.Collection }
108  
 109 func NewMongoRevoked(database *mongo.Database) *MongoRevoked {

⏺ Update(internal/auth/repository*mongo.go)
⎿  Added 10 lines
118 }
119 return n > 0, nil
120 }
121 +  
 122 +// Revoke inserts the jti as \_id. A duplicate means two logout
+s with the same token raced;  
 123 +// the token is revoked either way, so that is not an error.  
 124 +func (r \*MongoRevoked) Revoke(ctx context.Context, jti string, + userID bson.ObjectID, expiresAt time.Time) error {  
 125 + *, err := r.revoked.InsertOne(ctx, RevokedToken{JTI: jti, Us
+erID: userID, ExpiresAt: expiresAt})  
 126 + if err != nil && !mongo.IsDuplicateKeyError(err) {  
 127 + return fmt.Errorf("revoke access token: %w", err)  
 128 + }  
 129 + return nil  
 130 +}

⏺ Update(internal/auth/repository_mongo.go)
⎿  Added 1 line
4 "context"
5 "errors"
6 "fmt"
7 + "time"  
 8  
 9 "go.mongodb.org/mongo-driver/v2/bson"
10 "go.mongodb.org/mongo-driver/v2/mongo"

⏺ Model changes: RevokedToken, a Tokens pair for /refresh, and the token's
expiry on Claims (logout needs it for the revoked entry's TTL).

⏺ Update(internal/auth/model.go)
⎿  Added 2 lines, removed 1 line
17 }
18  
 19 // RefreshToken is the refresh_tokens document (PLAN.md section
1). Only the sha256 of the token is stored.
20 -// UsedAt and ReplacedBy are filled by rotation in phase 6.  
 20 +// UsedAt and ReplacedBy are set when the token is rotated. The + used token is kept, not deleted,  
 21 +// so a second use can be caught as theft.  
 22 type RefreshToken struct {
23 ID bson.ObjectID `bson:"_id"`
24 TokenHash string `bson:"tokenHash"`

⏺ Update(internal/auth/model.go)
⎿  Added 8 lines
30 ReplacedBy \*bson.ObjectID `bson:"replacedBy"`
31 }
32  
 33 +// RevokedToken is the revoked_tokens document. The TTL index o
+n expiresAt deletes it once the  
 34 +// access token would have expired anyway.  
 35 +type RevokedToken struct {  
 36 + JTI string `bson:"_id"`  
 37 + UserID bson.ObjectID `bson:"userId"`  
 38 + ExpiresAt time.Time `bson:"expiresAt"`  
 39 +}  
 40 +  
 41 // UserView is the API.md User shape.
42 type UserView struct {
43 ID string `json:"id"`

⏺ Update(internal/auth/model.go)
⎿  Added 13 lines, removed 6 lines
46 CreatedAt time.Time `json:"createdAt"`
47 }
48  
 49 -// Session is what register and login return.  
 49 +// Tokens is what /refresh returns.  
 50 +type Tokens struct {  
 51 + AccessToken string `json:"accessToken"`  
 52 + RefreshToken string `json:"refreshToken"`  
 53 +}  
 54 +  
 55 +// Session is what register and login return. Tokens is embedde
+d, so its fields sit at  
 56 +// the top level of the JSON next to "user".  
 57 type Session struct {
51 - AccessToken string `json:"accessToken"`  
 52 - RefreshToken string `json:"refreshToken"`  
 53 - User UserView `json:"user"`  
 58 + Tokens  
 59 + User UserView `json:"user"`  
 60 }
61  
 62 // Claims is what RequireAuth puts in the request context after
checking the access token.
63 type Claims struct {
58 - UserID bson.ObjectID  
 59 - JTI string  
 64 + UserID bson.ObjectID  
 65 + JTI string  
 66 + ExpiresAt time.Time // logout keeps the revoked jti until thi
+s time  
 67 }
68  
 69 func toView(u User) UserView {

⏺ Update(internal/auth/tokens.go)
⎿  Added 1 line, removed 1 line
47 if err != nil || claims.ID == "" {
48 return Claims{}, ErrTokenInvalid
49 }
50 - return Claims{UserID: userID, JTI: claims.ID}, nil  
 50 + return Claims{UserID: userID, JTI: claims.ID, ExpiresAt: clai
+ms.ExpiresAt.Time}, nil  
 51 }
52  
 53 // newRefreshToken returns 32 random bytes as base64url (sent t
o the client once) and its sha256 hex

⏺ Now the service: Refresh and Logout.

⏺ Update(internal/auth/service.go)
⎿  Added 2 lines
20 ErrTokenExpired = errors.New("access token expired")
21 ErrTokenInvalid = errors.New("access token invalid")
22 ErrTokenRevoked = errors.New("access token revoked")
23 + ErrRefreshInvalid = errors.New("refresh token unknown or
+expired")  
 24 + ErrRefreshReused = errors.New("refresh token used twice"
+)  
 25 )
26  
 27 // ValidationError lists broken field rules; the handler sends
it as 422 with "fields".

⏺ Update(internal/auth/service.go)
⎿  Added 81 lines, removed 8 lines
157 return toView(u), nil
158 }
159  
 160 -// newSession issues an access token and a refresh token. Each - login starts a new refresh  
 161 -// family; phase 6 rotation keeps the familyId along the chain
-.  
 162 -func (s *Service) newSession(ctx context.Context, u User) (Ses
-sion, error) {  
 160 +// Refresh swaps a refresh token for a new pair (PLAN.md secti
+on 8). Each refresh token works once.  
 161 +//  
 162 +// 1. Mark the token used, in one atomic step that only match
+es an unused, unexpired token.  
 163 +// 2. Matched: issue a new refresh token in the same family a
+nd a new access token.  
 164 +// 3. Not matched: if the token exists and was already used,
+someone is replaying it. Either the  
 165 +// user or a thief holds a copy, and we can't tell which,
+so every session of that user ends.  
 166 +func (s *Service) Refresh(ctx context.Context, raw string) (To
+kens, error) {  
 167 + if raw == "" {  
 168 + return Tokens{}, ErrRefreshInvalid  
 169 + }  
 170 now := s.now()
164 - access, err := signAccess(s.set.Secret, s.set.Issuer, u.ID,
-now, s.set.AccessTTL)  
 171 + hash := hashToken(raw)  
 172 + nextID := bson.NewObjectID() // known before the update, so
+step 1 can also set replacedBy  
 173 +  
 174 + old, err := s.refresh.Use(ctx, hash, now, nextID)  
 175 + if errors.Is(err, ErrRefreshNotFound) {  
 176 + return Tokens{}, s.refreshFailed(ctx, hash)  
 177 + }  
 178 if err != nil {
166 - return Session{}, fmt.Errorf("sign access token: %w", err)
179 + return Tokens{}, err  
 180 }
181 + // If anything below fails, the old token is already used an
+d the user has to log in again.  
 182 + // Accepted: the other order (insert first) can't work, as w
+e only learn the userId in step 1.  
 183 + return s.issue(ctx, nextID, old.UserID, old.FamilyID, now)  
 184 +}  
 185 +  
 186 +// refreshFailed tells a used token (theft: revoke all) from a
+n unknown or expired one.  
 187 +func (s *Service) refreshFailed(ctx context.Context, hash stri
+ng) error {  
 188 + t, err := s.refresh.ByHash(ctx, hash)  
 189 + if errors.Is(err, ErrRefreshNotFound) {  
 190 + return ErrRefreshInvalid  
 191 + }  
 192 + if err != nil {  
 193 + return err  
 194 + }  
 195 + if t.UsedAt == nil {  
 196 + return ErrRefreshInvalid // exists, unused, so it failed o
+n expiresAt  
 197 + }  
 198 + if err := s.refresh.DeleteUser(ctx, t.UserID); err != nil {
199 + return err  
 200 + }  
 201 + return ErrRefreshReused  
 202 +}  
 203 +  
 204 +// Logout ends one session: the access token's jti goes on the + revoked list (so it stops working  
 205 +// now, not in up to 15 minutes), and the refresh token's whol
+e family is deleted. Deleting the  
 206 +// family and not just the given token matters if a thief alre
+ady rotated a stolen copy: their  
 207 +// newer token is in the same family and dies too. An unknown, + empty or someone else's refresh  
 208 +// token is skipped without an error, since the client clears
+its tokens whatever we answer.  
 209 +func (s *Service) Logout(ctx context.Context, c Claims, rawRef
+resh string) error {  
 210 + if err := s.revoked.Revoke(ctx, c.JTI, c.UserID, c.ExpiresAt
+); err != nil {  
 211 + return err  
 212 + }  
 213 + if rawRefresh == "" {  
 214 + return nil  
 215 + }  
 216 + t, err := s.refresh.ByHash(ctx, hashToken(rawRefresh))  
 217 + if errors.Is(err, ErrRefreshNotFound) {  
 218 + return nil  
 219 + }  
 220 + if err != nil {  
 221 + return err  
 222 + }  
 223 + return s.refresh.DeleteFamily(ctx, c.UserID, t.FamilyID) //
+filters on userId, so only our own  
 224 +}  
 225 +  
 226 +// newSession is register and login: a new refresh family per
+login.  
 227 +func (s *Service) newSession(ctx context.Context, u User) (Ses
+sion, error) {  
 228 + tokens, err := s.issue(ctx, bson.NewObjectID(), u.ID, bson.N
+ewObjectID(), s.now())  
 229 + if err != nil {  
 230 + return Session{}, err  
 231 + }  
 232 + return Session{Tokens: tokens, User: toView(u)}, nil  
 233 +}  
 234 +  
 235 +// issue signs an access token and stores a new refresh token
+with the given id and family.  
 236 +func (s *Service) issue(ctx context.Context, id, userID, famil
+yID bson.ObjectID, now time.Time) (Tokens, error) {  
 237 + access, err := signAccess(s.set.Secret, s.set.Issuer, userID
+, now, s.set.AccessTTL)  
 238 + if err != nil {  
 239 + return Tokens{}, fmt.Errorf("sign access token: %w", err)
240 + }  
 241 raw, hash := newRefreshToken()
242 err = s.refresh.Insert(ctx, RefreshToken{
170 - ID: bson.NewObjectID(), TokenHash: hash, UserID: u.ID, Fam
-ilyID: bson.NewObjectID(),  
 243 + ID: id, TokenHash: hash, UserID: userID, FamilyID: familyI
+D,  
 244 CreatedAt: now, ExpiresAt: now.Add(s.set.RefreshTTL),
245 })
246 if err != nil {
174 - return Session{}, err  
 247 + return Tokens{}, err  
 248 }
176 - return Session{AccessToken: access, RefreshToken: raw, User: - toView(u)}, nil  
 249 + return Tokens{AccessToken: access, RefreshToken: raw}, nil  
 250 }
251  
 252 func normalizeEmail(email string) string {
⏺ Handler and routes:

⏺ Update(internal/auth/handler.go)
⎿  Added 31 lines
50 httpx.OK(w, session)
51 }
52  
 53 +func (h *Handler) refresh(w http.ResponseWriter, r *http.Reques
+t) {  
 54 + var body struct {  
 55 + RefreshToken string `json:"refreshToken"`  
 56 + }  
 57 + if err := httpx.DecodeJSON(w, r, &body); err != nil {  
 58 + httpx.WriteError(w, r, err)  
 59 + return  
 60 + }  
 61 + tokens, err := h.svc.Refresh(r.Context(), body.RefreshToken)
62 + if err != nil {  
 63 + httpx.WriteError(w, r, toHTTP(err))  
 64 + return  
 65 + }  
 66 + httpx.OK(w, tokens)  
 67 +}  
 68 +  
 69 +func (h *Handler) logout(w http.ResponseWriter, r *http.Request
+) {  
 70 + var body struct {  
 71 + RefreshToken string `json:"refreshToken"`  
 72 + }  
 73 + if err := httpx.DecodeJSON(w, r, &body); err != nil {  
 74 + httpx.WriteError(w, r, err)  
 75 + return  
 76 + }  
 77 + if err := h.svc.Logout(r.Context(), ClaimsFrom(r.Context()),
+body.RefreshToken); err != nil {  
 78 + httpx.WriteError(w, r, toHTTP(err))  
 79 + return  
 80 + }  
 81 + httpx.NoContent(w)  
 82 +}  
 83 +  
 84 func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
85 user, err := h.svc.Me(r.Context(), ClaimsFrom(r.Context()).Us
erID)
86 if err != nil {

⏺ Update(internal/auth/handler.go)
⎿  Added 4 lines
109 return httpx.NewError(http.StatusUnauthorized, httpx.CodeT
okenRevoked, "Access token revoked")
110 case errors.Is(err, ErrTokenInvalid):
111 return httpx.NewError(http.StatusUnauthorized, httpx.CodeU
nauthorized, "Invalid access token")
112 + case errors.Is(err, ErrRefreshInvalid):  
 113 + return httpx.NewError(http.StatusUnauthorized, httpx.CodeU
+nauthorized, "Invalid or expired refresh token")  
 114 + case errors.Is(err, ErrRefreshReused):  
 115 + return httpx.NewError(http.StatusUnauthorized, httpx.CodeT
+okenReused, "Refresh token already used; all sessions were sig
+ned out")  
 116 }
117 return err
118 }

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && cat
internal/platform/httpx/respond.go internal/platform/httpx/decode.go)
⎿  (No output)

⏺ Routes, with a separate limiter for /refresh:

⏺ Write(internal/auth/routes.go)
⎿  Added 21 lines, removed 11 lines
15 r := chi.NewRouter()
16  
 17 // One limiter shared by /register and /login only: 10 reques
ts per minute per IP across both,
18 - // to slow down password guessing. /me is not limited: the cl
-ient calls it on every protected  
 19 - // page and in the header, and it already needs a valid token
-. /refresh (phase 6) gets its own,  
 20 - // higher limit, since every open tab refreshes every 15 minu
-tes.  
 21 - // The key is r.RemoteAddr, which middleware.ClientIP has alr
-eady set to the real client IP.  
 22 - // Not httprate.KeyByRealIP: it trusts headers the client can - fake.  
 23 - limit := httprate.Limit(10, time.Minute,  
 18 + // to slow down password guessing. /me and /logout are not li
+mited: they already need a valid  
 19 + // token, and /me runs on every protected page and in the hea
+der.  
 20 + login := perIP(10)  
 21 + // /refresh has its own limiter, so failed logins from the sa
+me IP (a shared office or home  
 22 + // network) never block a signed-in tab from refreshing. 60 i
+s generous: each tab refreshes  
 23 + // about once per 15 minutes. It is still limited because a r
+equest costs db writes.  
 24 + refresh := perIP(60)  
 25 +  
 26 + r.With(login).Post("/register", h.register)  
 27 + r.With(login).Post("/login", h.login)  
 28 + r.With(refresh).Post("/refresh", h.refresh)  
 29 + r.With(h.RequireAuth).Post("/logout", h.logout)  
 30 + r.With(h.RequireAuth).Get("/me", h.me)  
 31 + return r  
 32 +}  
 33 +  
 34 +// perIP allows n requests per minute per client IP. Each call
+makes its own counter.  
 35 +// The key is r.RemoteAddr, which middleware.ClientIP has alrea
+dy set to the real client IP.  
 36 +// Not httprate.KeyByRealIP: it trusts headers the client can f
+ake.  
 37 +func perIP(n int) func(http.Handler) http.Handler {  
 38 + return httprate.Limit(n, time.Minute,  
 39 httprate.WithKeyFuncs(func(r \*http.Request) (string, error)
{
40 return httprate.CanonicalizeIP(r.RemoteAddr), nil
41 }),
...
43 httpx.WriteError(w, r, httpx.NewError(http.StatusTooManyR
equests, httpx.CodeRateLimited, "Too many attempts, try again i
n a minute"))
44 }),
45 )
31 -  
 32 - r.With(limit).Post("/register", h.register)  
 33 - r.With(limit).Post("/login", h.login)  
 34 - r.With(h.RequireAuth).Get("/me", h.me)  
 35 - return r  
 46 }

⏺ Now the in-memory fakes and the tests.

⏺ Update(internal/auth/service*test.go)
⎿  Added 50 lines
64 return nil
65 }
66  
 67 +// Use holds the lock for both the check and the write, like t
+he atomic FindOneAndUpdate.  
 68 +func (m \*memRefresh) Use(* context.Context, hash string, now t
+ime.Time, replacedBy bson.ObjectID) (RefreshToken, error) {  
 69 + m.mu.Lock()  
 70 + defer m.mu.Unlock()  
 71 + for i, t := range m.tokens {  
 72 + if t.TokenHash == hash && t.UsedAt == nil && t.ExpiresAt.A
+fter(now) {  
 73 + m.tokens[i].UsedAt, m.tokens[i].ReplacedBy = &now, &repl
+acedBy  
 74 + return t, nil  
 75 + }  
 76 + }  
 77 + return RefreshToken{}, ErrRefreshNotFound  
 78 +}  
 79 +  
 80 +func (m *memRefresh) ByHash(* context.Context, hash string) (R
+efreshToken, error) {  
 81 + m.mu.Lock()  
 82 + defer m.mu.Unlock()  
 83 + for *, t := range m.tokens {  
 84 + if t.TokenHash == hash {  
 85 + return t, nil  
 86 + }  
 87 + }  
 88 + return RefreshToken{}, ErrRefreshNotFound  
 89 +}  
 90 +  
 91 +func (m *memRefresh) DeleteUser(_ context.Context, userID bson
+.ObjectID) error {  
 92 + return m.deleteWhere(func(t RefreshToken) bool { return t.Us
+erID == userID })  
 93 +}  
 94 +  
 95 +func (m \*memRefresh) DeleteFamily(_ context.Context, userID, f
+amilyID bson.ObjectID) error {  
 96 + return m.deleteWhere(func(t RefreshToken) bool { return t.Us
+erID == userID && t.FamilyID == familyID })  
 97 +}  
 98 +  
 99 +func (m *memRefresh) deleteWhere(match func(RefreshToken) bool
+) error {  
 100 + m.mu.Lock()  
 101 + defer m.mu.Unlock()  
 102 + kept := m.tokens[:0]  
 103 + for \_, t := range m.tokens {  
 104 + if !match(t) {  
 105 + kept = append(kept, t)  
 106 + }  
 107 + }  
 108 + m.tokens = kept  
 109 + return nil  
 110 +}  
 111 +  
 112 type memRevoked struct {
113 mu sync.Mutex
114 ids map[string]bool
...
120 return m.ids[jti], nil
121 }
122  
 123 +func (m *memRevoked) Revoke(_ context.Context, jti string, _ b
+son.ObjectID, \_ time.Time) error {  
 124 + m.revoke(jti)  
 125 + return nil  
 126 +}  
 127 +  
 128 func (m \*memRevoked) revoke(jti string) {
129 m.mu.Lock()
130 defer m.mu.Unlock()

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && cat >>
internal/auth/service_test.go <<'EOF'…)
⎿  (No output)

⏺ Now route-level tests for the status codes and the separate refresh limit.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && cat >>
internal/auth/middleware_test.go <<'EOF'…)
⎿  (No output)

⏺ All 17 pass. Next: the race detector, then a live run against eneba_dev.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && go test -race
./... -count=1 2>&1 | tail -15; ls -a | grep -E "^\.env"; lsof -i :8080
-sTCP:LISTEN |…)
⎿  (No output)

⏺ A server is already on :8080, probably the owner's air, which rebuilds on
save. I'll check that it serves the new routes, then run the flow.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && ps -o command=
-p 78423; curl -s -X POST localhost:8080/api/auth/refresh -d
'{"refreshToken":"nope"}'…)
⎿  (No output)

⏺ That binary is an older build, and I won't touch the owner's process. I'll
start my own copy on port 8090.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && go build -o
/private/tmp/claude-501/-Users-onlyeugene-Desktop-Eneba-Marketplace/7a2a81
11-9216-4a19-8a…)
⎿  (No output)

⏺ Bash(B=localhost:8090/api/auth; E="phase6-$(date +%s)@example.com"
j() { python3 -c "import sys,json;print(json.load(sys.stdin)['$1'])"; }…)
⎿  (No output)

⏺ The live flow matches API.md. Now I'll confirm the TTL indexes exist on
eneba_dev and that expiresAt is stored as a real date, since TTL ignores
strings.

Searched for 1 pattern (ctrl+o to expand)

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && mkdir -p
./tmp/ttlcheck && cat > ./tmp/ttlcheck/main.go <<'EOF'
package main…)
⎿  (No output)

⏺ Both TTL indexes are in place, and expiresAt is a BSON date. I'll stop my
server and update the docs.

⏺ Bash(pkill -f "scratchpad/api$"; sleep 0.3; lsof -i :8090 -sTCP:LISTEN || echo
"8090 stopped")
⎿  (No output)

⏺ Update(~/Desktop/Eneba Marketplace/API.md)
⎿  Added 2 lines, removed 2 lines
111 - Password: 8–72 bytes (bcrypt ignores bytes past 72), at leas
t one digit or symbol. Email lowercased and trimmed. `name` op
tional, at most 50 chars, stored as `""` when left out.
112 - Broken rules → `422 VALIDATION_FAILED` with `fields` (e.g. `
           { "email": "is not a valid email", "password": "must contain a
            digit or a symbol" }`). Email already registered → `409 EMAIL
           _TAKEN`.
113 - Login returns the same `401 UNAUTHORIZED "Invalid email or p
           assword"` for unknown email and wrong password, so it doesn't
reveal which emails exist. An empty email or password is a `42
           2`.
114 -- Refresh: unknown or expired token → `401 UNAUTHORIZED`. Alre
-ady-used token → `401 TOKEN_REUSED` and **every refresh token
-for that user is deleted** (all sessions must log in again).  
 115 -- Logout: adds the access token's `jti` to `revoked_tokens`, d
-eletes the given refresh token. The client clears local tokens - even if this call fails.  
 114 +- Refresh: each refresh token works once; the response carries + a new pair. Unknown, empty or expired token → `401 UNAUTHORIZ
          +ED`. Already-used token → `401 TOKEN_REUSED` and **every refre
+sh token for that user is deleted** (all sessions must log in
+again).  
 115 +- Logout: adds the access token's `jti` to `revoked_tokens` (t
+hat access token stops working at once) and deletes the given
+refresh token along with every token rotated from the same log
+in. Other logins of the same user stay signed in. An unknown,
+empty or someone else's `refreshToken` is skipped, still `204`
+. The client clears local tokens even if this call fails.  
 116 - Middleware errors: `401 TOKEN_EXPIRED` (client should refres
h), `401 TOKEN_REVOKED` / `401 UNAUTHORIZED` (client should al
so try refresh once; if that fails, log out). Missing header,
bad signature, wrong `alg` or issuer → `UNAUTHORIZED`.
117 - `/register` and `/login` share one rate limit per IP: 10 req
uests per minute across both, then `429 RATE_LIMITED`. `/refre
           sh` has its own, higher limit (set in phase 6). `/me` and `/lo
           gout` are not rate limited.
118 - Malformed JSON body → `400 BAD_REQUEST`.

⏺ Update(~/Desktop/Eneba Marketplace/API.md)
⎿  Added 1 line, removed 1 line
114 - Refresh: each refresh token works once; the response carries
a new pair. Unknown, empty or expired token → `401 UNAUTHORIZ
           ED`. Already-used token → `401 TOKEN_REUSED` and **every refre
sh token for that user is deleted** (all sessions must log in
again).
115 - Logout: adds the access token's `jti` to `revoked_tokens` (t
hat access token stops working at once) and deletes the given
refresh token along with every token rotated from the same log
in. Other logins of the same user stay signed in. An unknown,
empty or someone else's `refreshToken` is skipped, still `204`
. The client clears local tokens even if this call fails.
116 - Middleware errors: `401 TOKEN_EXPIRED` (client should refres
h), `401 TOKEN_REVOKED` / `401 UNAUTHORIZED` (client should al
so try refresh once; if that fails, log out). Missing header,
bad signature, wrong `alg` or issuer → `UNAUTHORIZED`.
117 -- `/register` and `/login` share one rate limit per IP: 10 req
-uests per minute across both, then `429 RATE_LIMITED`. `/refre
          -sh` has its own, higher limit (set in phase 6). `/me` and `/lo
          -gout` are not rate limited.  
 117 +- `/register` and `/login` share one rate limit per IP: 10 req
+uests per minute across both, then `429 RATE_LIMITED`. `/refre
          +sh` has its own limit: 60 per minute per IP, counted apart fro
+m login, so failed logins never block a refresh. `/me` and `/l
          +ogout` are not rate limited.  
 118 - Malformed JSON body → `400 BAD_REQUEST`.
119  
 120 ## Catalog and search
⏺ Update(~/Desktop/Eneba Marketplace/API.md)
⎿  Added 1 line, removed 1 line
19 | Status | Code examples | When |
20 |---|---|---|
21 | 400 | `BAD_REQUEST` | Malformed JSON, bad query param |
22 -| 401 | `UNAUTHORIZED`, `TOKEN_EXPIRED`, `TOKEN_REVOKED` | Miss
-ing/invalid/expired/revoked access token, bad credentials |  
 22 +| 401 | `UNAUTHORIZED`, `TOKEN_EXPIRED`, `TOKEN_REVOKED`, `TOKE
         +N_REUSED` | Missing/invalid/expired/revoked access token, bad c
+redentials, refresh token used twice |  
 23 | 402 | `PAYMENT_DECLINED` | Simulated card declined |
24 | 403 | `FORBIDDEN` | Order belongs to another user |
25 | 404 | `NOT_FOUND` | |

⏺ Update(~/Desktop/Eneba Marketplace/PLAN.md)
⎿  Added 1 line, removed 1 line
726 | 3 ✅ | Catalog API: list with filters/sort/pagination (no `q
           `), filters, product detail + offers, cover route | curl each;
open a `cover.svg` in the browser |
727 | 4 ✅ | Atlas Search: create index from JSON, `q` on `/api/pr
           oducts`, `/api/search/autocomplete` | Typo query and partial q
uery return the right product |
728 | 5 ✅ | Auth part 1: users, register/login, JWT, middleware,
`/me`, rate limit + tests for expired/revoked | `go test ./int
           ernal/auth/...` green; curl register → me |
729 -| 6 | Auth part 2: refresh rotation, reuse detection, logout,
-TTL indexes + tests | Tests for rotation and reuse green |  
 729 +| 6 ✅ | Auth part 2: refresh rotation, reuse detection, logou
+t, TTL indexes + tests | Tests for rotation and reuse green |
730 | 7 | Homepage: layout endpoint, registry, 6 block services, `
           ?fail=1` | Each block id returns data; `?fail=1` returns 503 |
731 | 8 | Cart API: quote, get, put, delete, merge | curl flow: me
rge → put → get shows live prices |
732 | 9 | Orders: reserve transaction with expiry-aware claim, sim
ulated pay, cleanup job, list/detail, `claim_integration_test.
           go` | `go test ./internal/orders/...` green (unit only); `go t
           est -tags integration ./internal/orders/...` → all 4 integrati
on tests green; decline puts stock back and recounts immediate
ly |

⏺ Update(~/Desktop/Eneba Marketplace/PLAN.md)
⎿  Added 3 lines
765 | Auth rate limit | One `httprate` limiter shared by `/registe
           r` and `/login`: 10 per minute per IP across both, not per rou
te. `/me` is not limited: the client calls it on every protect
ed page and in the header, and it already needs a valid token.
`/refresh` (phase 6) gets its own, higher limiter, because ev
ery open tab refreshes every 15 minutes and must not be blocke
d by failed logins from the same IP. The key is `r.RemoteAddr`
after our `ClientIP` middleware, not httprate's `KeyByRealIP`
, which trusts headers the client sets. |
766 | Login timing | For an unknown email, login still runs one bc
rypt compare against a dummy hash, so the response time doesn'
t reveal whether the email has an account. |
767 | Auth errors in the service | The service returns plain error
s (`ErrTokenExpired`, `ValidationError`, …) and `auth.toHTTP`
maps them to API.md codes, so the service never imports `net/h
           ttp`. |
768 +| Logout deletes the refresh family | Logout deletes every ref
+resh token with the given token's `familyId` (and the caller's + `userId`), not just that one token. If a thief rotated a stol
+en copy first, their newer token is in the same family, so log
+out ends it too. Other logins (other families) stay signed in. + Same cost: one lookup, one `DeleteMany`. |  
 769 +| Refresh write order | `/refresh` marks the old token used (a
+nd sets `replacedBy` to an id picked in advance) before insert
+ing the new one, since the `userId` is only known after that f
+irst atomic step. If the insert fails, the user has to log in
+again. Rare, and safe. |  
 770 +| `/refresh` rate limit | 60 per minute per IP, own counter. E
+ach tab refreshes about once per 15 minutes, so this only stop
+s floods. |  
 771 | `golang.org/x/crypto` version | Pinned at v0.54. Newer versi
ons need Go 1.26, which would raise the `go` line in go.mod an
d the toolchain the Dockerfile needs. |
772  
 773 ## 12. Future improvements

⏺ Update(~/Desktop/Eneba Marketplace/PLAN.md)
⎿  Added 1 line, removed 1 line
761 | Search `q` length | Trimmed, capped at 100 chars (`400` abov
e). Fuzzy matching cost grows with every term, and no product
name is that long. |
762 | Where `Searcher` lives | `products.Searcher` interface, impl
emented by `search.Service`. `search` imports `products` (for
`ListQuery`, `ToCard`), so the interface has to sit in `produc
           ts` to avoid an import cycle. |
763 | Unknown errors | `httpx.WriteError` logs them and returns `5
           00 INTERNAL "Something went wrong"`, so driver messages never
reach the browser. |
764 -| Refresh token in phase 5 | Register and login already return - and store a refresh token (hash only, new `familyId` per logi
-n), because API.md promises it. `/refresh` rotation, reuse det
-ection and logout come in phase 6. |  
 764 +| Refresh token in phase 5 | Register and login already return + and store a refresh token (hash only, new `familyId` per logi
+n), because API.md promises it. `/refresh` rotation, reuse det
+ection and logout were added in phase 6. |  
 765 | Auth rate limit | One `httprate` limiter shared by `/registe
           r` and `/login`: 10 per minute per IP across both, not per rou
te. `/me` is not limited: the client calls it on every protect
ed page and in the header, and it already needs a valid token.
`/refresh` (phase 6) gets its own, higher limiter, because ev
ery open tab refreshes every 15 minutes and must not be blocke
d by failed logins from the same IP. The key is `r.RemoteAddr`
after our `ClientIP` middleware, not httprate's `KeyByRealIP`
, which trusts headers the client sets. |
766 | Login timing | For an unknown email, login still runs one bc
rypt compare against a dummy hash, so the response time doesn'
t reveal whether the email has an account. |
767 | Auth errors in the service | The service returns plain error
s (`ErrTokenExpired`, `ValidationError`, …) and `auth.toHTTP`
maps them to API.md codes, so the service never imports `net/h
           ttp`. |

⏺ PLAN.md section 8 still says replacedBy is set after the insert. I'll make it
match the code.

⏺ Update(~/Desktop/Eneba Marketplace/PLAN.md)
⎿  Added 2 lines, removed 2 lines
651  
 652 **Refresh flow** (`POST /api/auth/refresh`):
653 1. `h = sha256(token)`.
654 -2. `FindOneAndUpdate({ tokenHash: h, usedAt: null, expiresAt: 
          -{ $gt: now } }, { $set: { usedAt: now } })`. Atomic, so two pa
-rallel calls with the same token can't both succeed.  
 655 -3. Matched → insert new token (same `familyId`), set `replaced
          -By` on the old one, issue new access token, return the pair.  
 654 +2. Pick the new token's `_id`, then `FindOneAndUpdate({ tokenH
          +ash: h, usedAt: null, expiresAt: { $gt: now } }, { $set: { use
          +dAt: now, replacedBy: newId } })`. Atomic, so two parallel cal
+ls with the same token can't both succeed.  
 655 +3. Matched → insert the new token with that `_id` (same `famil
          +yId`), issue a new access token, return the pair.  
 656 4. Not matched → `FindOne({ tokenHash: h })`. If it exists and
`usedAt` is set → **reuse**: `DeleteMany({ userId })`, return
`401 TOKEN_REUSED`. Otherwise → `401 UNAUTHORIZED`.
657  
 658 The used token is marked, not deleted, so a later replay can s
till be recognized. The TTL index removes it after 30 days.

⏺ Phase 6 is done and marked ✅ in PLAN.md. Sign-in sessions now renew
themselves with refresh tokens, and a token that gets used twice signs the
user out everywhere. Logout ends the session at once. All 17 auth tests pass,
including under Go's race checker, and I ran the full flow against eneba_dev
with the expected results.

What changed

- POST /api/auth/refresh: trades a refresh token for a new pair. Each token
  works only once: the step that checks it and marks it used is a single
  database operation, so two requests with the same token can't both succeed.
  The old token stays in the database, marked used, so a second use can be
  caught.
- Reuse: if someone presents a token that was already used, the server deletes
  every refresh token for that user and returns 401 TOKEN_REUSED.
- POST /api/auth/logout (needs the access token): blocks that access token at
  once and deletes the refresh token. It also deletes every token renewed from
  the same login, so if a thief renewed a stolen copy, their token dies too.
  The user's other logins stay signed in.
- Rate limit: /refresh gets its own limit of 60 per minute per IP. Failed
  logins from the same network never block a refresh.
- Auto-delete (TTL) indexes: these already existed from phase 2. I checked
  that both are live on eneba_dev and that expiresAt is stored as a date,
  which they need in order to delete anything.
- API.md and PLAN.md are updated, and three new decisions are logged in
  PLAN.md section 11.

Two calls for you to confirm:

- Logout deletes the whole login's tokens. API.md said logout deletes only
  "the given refresh token". I widened it because it costs one extra lookup
  and closes the stolen-copy case above. Say if you'd rather keep the narrow
  version.
- Failed renewal signs the user out. /refresh marks the old token used before
  it saves the new one. If that save fails, the user has to log in again. It's
  rare and does no harm, and it's noted in PLAN.md.

Assignment check

This covers Authentication ("create an account and sign in") on the server
side. Still missing: the login/register pages, the account menu, and guarding
checkout and orders in the client (phase 14), plus the client's token refresh
helper (phase 10).

How to check

Tests:
cd server && go test ./internal/auth/... -v -count=1
The new tests are TestRefreshRotation, TestRefreshReuseSignsOutEverySession,
TestRefreshParallelOnlyOneWins, TestRefreshUnknownOrExpired, TestLogout,
TestLogoutIgnoresOtherUsersToken, TestRefreshAndLogoutRoutes and
TestRefreshHasItsOwnRateLimit.

Restart the API first. Your air on :8080 was serving an old build:
/api/auth/refresh returned 404. Stop it and run air again, or use go run
./cmd/api.

Postman

Base URL http://localhost:8080. For each POST, set Body → raw → JSON.

1. Register: POST /api/auth/register
   { "email": "p6@example.com", "password": "hunter22!", "name": "Eugene" }
   → 201. Save accessToken as A1 and refreshToken as R1.
2. Refresh: POST /api/auth/refresh
   { "refreshToken": "<R1>" }
   → 200 { accessToken, refreshToken }. Save them as A2 and R2.
3. Me with the new token: GET /api/auth/me, Authorization → Bearer Token = A2
   → 200 { user }.
4. Reuse R1: send request 2 again, unchanged → 401 TOKEN_REUSED.
5. R2 is now dead too: POST /api/auth/refresh with { "refreshToken": "<R2>" }
   → 401 UNAUTHORIZED. The reuse in step 4 signed out every session.
6. Log in again: POST /api/auth/login
   { "email": "p6@example.com", "password": "hunter22!" }
   → 200. Save the tokens as A3 and R3.
7. Logout: POST /api/auth/logout, Bearer = A3
   { "refreshToken": "<R3>" }
   → 204, empty body.
8. Me after logout: GET /api/auth/me, Bearer = A3 → 401 TOKEN_REVOKED.
9. Refresh after logout: POST /api/auth/refresh with R3 → 401 UNAUTHORIZED.
10. Logout without a token: request 7 with Authorization set to No Auth → 401
    UNAUTHORIZED.

Optional, separate rate limits: send 11 bad logins in a row; the 11th returns
429 RATE_LIMITED. A refresh with a fresh, valid refresh token from the same IP
still returns 200.

In Atlas (eneba_dev), refresh_tokens shows the used R1 with usedAt set, until
the reuse in step 4 deletes it. revoked_tokens holds A3's jti, with an
expiresAt 15 minutes out. The TTL index deletes it about a minute after that.

❯ keep both. logout deleting the whole login's tokens is the safer choice, and
a failed renewal
asking the user to log in again is fine.

⏺ Good, both stay as built. Nothing to change: PLAN.md section 11 already
records both decisions, and API.md already describes logout as deleting the
whole login's tokens.

Phase 6 is finished. Phase 7 (homepage blocks) is next whenever you're ready.

❯ why do we need both a blacklist for access tokens and deleting refresh tokens
on logout? wouldn't one be enough?

⏺ No, each one stops a different token, and neither covers the other.

Deleting the refresh token alone leaves the access token working. An access
token is a signed JWT, and the server checks it by verifying the signature. It
never looks it up in the database. After logout, the access token stays valid
until it expires, which can take up to 15 minutes. Anyone holding a copy,
such as another script on the page or a stolen copy, can keep calling the API
until then.

Blocking the access token alone leaves the refresh token working. The refresh
token can mint a fresh access token at any time for 30 days. Someone holding
it just calls /refresh and gets a new access token whose jti isn't on the
blocked list. Logout would end nothing.

┌───────────────────┬───────────────┬────────────────────────────────────┐
│ Logout does │ Access token │ Refresh token │
├───────────────────┼───────────────┼────────────────────────────────────┤
│ Delete refresh │ works ≤ 15 │ dead │
│ only │ min │ │
├───────────────────┼───────────────┼────────────────────────────────────┤
│ Block access only │ dead │ works 30 days, mints new access │
│ │ │ tokens │
├───────────────────┼───────────────┼────────────────────────────────────┤
│ Both │ dead │ dead │
└───────────────────┴───────────────┴────────────────────────────────────┘

The trade-off: the blocked list isn't free. Every protected request does one
extra database read in RequireAuth to check the jti (service.go:131-144
(server/internal/auth/service.go#L131-L144)). Many systems skip it: they
delete the refresh token and accept that the access token lives out its short
expiry. That would be a fair choice here too, since 15 minutes is a small
window. We kept the list because the plan (section 8) chose it. It means
logout takes effect at once, and the cost at this scale is one read on the
indexed \_id.

Deleting the refresh token is the part you can't drop. Without it, logout does
nothing that lasts. The blocked list only closes the last ≤ 15 minutes.

❯ walk me through what happens when two tabs send the same refresh token at the
same moment.

⏺ Here's the timeline on the server. Tab 1 holds token A and so does Tab 2; both
send POST /refresh { A } within the same few milliseconds.

1. Both reach the atomic step. Each request runs Use, a single
   FindOneAndUpdate with the filter { tokenHash: A, usedAt: null, expiresAt > now
   } (repository_mongo.go (server/internal/auth/repository_mongo.go)). Mongo
   applies writes to one document one at a time, so only one of them can match:

- Tab 1 wins. A is marked usedAt: now, replacedBy: B_id.
- Tab 2 loses. By the time its update runs, usedAt is no longer null, so it
  matches nothing.

2. The winner issues new tokens. Tab 1 signs a new access token and inserts
   refresh token B (same family).

3. The loser looks closer. Tab 2 calls ByHash(A), finds A with usedAt set, and
   treats it as theft. It can't tell a second tab from an attacker. It runs
   DeleteUser, deleting every refresh token of that user, and returns 401
   TOKEN_REUSED.

4. The outcome depends on which came first, step 2's insert or step 3's
   delete:

- Insert before delete (the usual case): B is deleted too. Tab 2 is signed out
  now. Tab 1's access token works until it expires, then its refresh with B
  fails, and it's signed out as well. The user logs in again.
- Delete before insert (a narrow window): B is written after the cleanup, so
  it survives. Tab 1 stays signed in, and only Tab 2 is out.

What this means

For a real user, this is a false alarm. Two of their own tabs set off theft
detection and log them out. The server can't avoid that: from its side, the
two requests look the same. The fix belongs in the client, and the plan has it
for phase 10 (PLAN.md section 8):

- Within one tab, refreshOnce() shares one promise, so five 401s cause one
  refresh.
- Across tabs, the refresh runs inside
  navigator.locks.request("auth-refresh"). Tab 2 waits for the lock, then
  reads localStorage again. It sees that Tab 1 already stored B and uses B
  instead of sending A. The server never sees A twice.

Until phase 10 is built, nothing stops two tabs from racing, but nothing calls
/refresh yet either.

For a theft, the narrow window is a small gap. Say the attacker's request is
Tab 1 and the user's is Tab 2. Usually the attacker's new token B is deleted,
which is the point of reuse detection. In the delete-before-insert case, B
survives and the attacker keeps a session for up to 30 days, while the user is
signed out. Pulling that off means sending the stolen token within
milliseconds of the real user's refresh. That's unlikely, but it is a hole in
the guarantee.

A possible fix, about four lines in Refresh: after inserting B, check that A
still exists with ByHash. If it's gone, a reuse cleanup ran in between, so
delete B and return TOKEN_REUSED. It costs one extra read per refresh.
TestRefreshParallelOnlyOneWins today checks only that exactly one refresh
wins, not what happens to B afterwards.

Should I add that check and a test for it, or note the gap in PLAN.md section
12 and move on? Either works: the client lock prevents the false alarm, and
the theft window is very narrow.

❯ add the check and a test for it. it's a security gap and it's small.

⏺ I'll write the test first and run it against the current code, so it shows the
gap before the fix.

⏺ Update(internal/auth/service*test.go)
⎿  Added 7 lines
55 type memRefresh struct {
56 mu sync.Mutex
57 tokens []RefreshToken
58 + // beforeInsert, if set, runs once just before the next Inser
+t. Tests use it to make  
 59 + // another request land between Refresh's atomic Use and its
+Insert.  
 60 + beforeInsert func()  
 61 }
62  
 63 func (m \*memRefresh) Insert(* context.Context, t RefreshToken)
error {
64 + if h := m.beforeInsert; h != nil {  
 65 + m.beforeInsert = nil  
 66 + h()  
 67 + }  
 68 m.mu.Lock()
69 defer m.mu.Unlock()
70 m.tokens = append(m.tokens, t)

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && cat >>
internal/auth/service_test.go <<'EOF'…)
⎿  (No output)

⏺ The test fails before the fix as expected, though part of the "2 tokens left"
count was my own refresh adding a token. I'll recount first, then apply the
fix.

⏺ Update(internal/auth/service*test.go)
⎿  Added 4 lines, removed 3 lines
475 if err == nil {
476 t.Error("winner must not get a pair once its chain was wip
ed")
477 }
478 + // Counted before the next check, which would itself rotate
+and add a token.  
 479 + if n := len(env.refresh.tokens); n != 0 {  
 480 + t.Errorf("%d refresh tokens left for the user, want 0", n)
481 + }  
 482 if won.RefreshToken != "" {
483 if *, err := env.svc.Refresh(ctx, won.RefreshToken); err =
= nil {
484 t.Error("winner's new refresh token must not work")
485 }
486 }
483 - if n := len(env.refresh.tokens); n != 0 {  
 484 - t.Errorf("%d refresh tokens left for the user, want 0", n)
485 - }  
 487 }

⏺ Update(internal/auth/service.go)
⎿  Added 18 lines, removed 1 line
180 }
181 // If anything below fails, the old token is already used an
d the user has to log in again.
182 // Accepted: the other order (insert first) can't work, as w
e only learn the userId in step 1.
183 - return s.issue(ctx, nextID, old.UserID, old.FamilyID, now)  
 183 + tokens, err := s.issue(ctx, nextID, old.UserID, old.FamilyID
+, now)  
 184 + if err != nil {  
 185 + return Tokens{}, err  
 186 + }  
 187 +  
 188 + // A replay of the old token may have run the reuse cleanup
+between our Use and our insert.  
 189 + // Then the new token was written after the cleanup and woul
+d survive it, so a thief who won  
 190 + // the race would keep a session. Used tokens are kept, neve
+r deleted, so if the old one is  
 191 + // gone, a cleanup (or a logout of this family) ran: delete
+the new token too and refuse.  
 192 + if \_, err := s.refresh.ByHash(ctx, hash); errors.Is(err, Err
+RefreshNotFound) {  
 193 + if err := s.refresh.DeleteFamily(ctx, old.UserID, old.Fami
+lyID); err != nil {  
 194 + return Tokens{}, err  
 195 + }  
 196 + return Tokens{}, ErrRefreshInvalid  
 197 + } else if err != nil {  
 198 + return Tokens{}, err  
 199 + }  
 200 + return tokens, nil  
 201 }
202  
 203 // refreshFailed tells a used token (theft: revoke all) from a
n unknown or expired one.

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && gofmt -l
./internal; go vet ./... && go test -race ./internal/auth/... -count=1 -v
2>&1 | grep -E "^(…)
⎿  (No output)

⏺ The new test passes, but TestRefreshParallelOnlyOneWins now fails, which is
correct since it wrongly expected one winner; I'll update it to require at
most one pair issued and no surviving token.

⏺ Update(internal/auth/service_test.go)
⎿  Added 3 lines, removed 1 line
368 }
369 }
370  
 371 -// Many parallel refreshes with one token: exactly one gets a
-new pair.  
 371 +// Many parallel refreshes with one token: at most one gets a
+new pair, and since the losers  
 372 +// count as reuse, no refresh token survives. Zero winners is
+fine: if a loser's cleanup lands  
 373 +// before the winner's final check, the winner is refused too. + Which one happens is up to timing.  
 374 func TestRefreshParallelOnlyOneWins(t \*testing.T) {
375 env := newTestEnv(t)
376 a := env.register(t, "eugene@example.com")

⏺ Update(internal/auth/service_test.go)
⎿  Added 5 lines, removed 2 lines
388 })
389 }
390 wg.Wait()
391 - if wins != 1 {  
 392 - t.Errorf("%d refreshes succeeded, want exactly 1", wins)  
 391 + if wins > 1 {  
 392 + t.Errorf("%d refreshes succeeded, want at most 1", wins)  
 393 }
394 + if n := len(env.refresh.tokens); n != 0 {  
 395 + t.Errorf("%d refresh tokens survived the race, want 0", n)
396 + }  
 397 }
398  
 399 func TestRefreshUnknownOrExpired(t \*testing.T) {

⏺ Bash(cd "/Users/onlyeugene/Desktop/Eneba Marketplace/server" && go test -race
./internal/auth/... -count=50 2>&1 | tail -5 && go test ./... -count=1
2>&1 | grep -v "…)
⎿  (No output)

⏺ All tests pass 50 runs in a row under the race checker. I'll add the new step
to PLAN.md section 8.

⏺ Update(~/Desktop/Eneba Marketplace/PLAN.md) ⎿  Added 1 line
654 2. Pick the new token's `_id`, then `FindOneAndUpdate({ tokenH
           ash: h, usedAt: null, expiresAt: { $gt: now } }, { $set: { use
           dAt: now, replacedBy: newId } })`. Atomic, so two parallel cal
ls with the same token can't both succeed.
655 3. Matched → insert the new token with that `_id` (same `famil
           yId`), issue a new access token, return the pair.
656 4. Not matched → `FindOne({ tokenHash: h })`. If it exists and
`usedAt` is set → **reuse**: `DeleteMany({ userId })`, return
`401 TOKEN_REUSED`. Otherwise → `401 UNAUTHORIZED`.
657 +5. After step 3's insert, check the old token still exists. Us
+ed tokens are never deleted except by a reuse cleanup or a log
+out of that family, so if it's gone, one of those ran between
+step 2 and the insert, and the new token was written after it. + Delete the family and return `401 UNAUTHORIZED`. Without this
+, a thief whose replay won the race would keep a new token tha
+t the cleanup missed. Cost: one indexed read per refresh. Test
+: `TestRefreshReuseCleanupBetweenUseAndInsert`.  
 658  
 659 The used token is marked, not deleted, so a later replay can s
till be recognized. The TTL index removes it after 30 days.
660

⏺ The gap is closed. A replay that lands between a refresh's two database steps
now leaves no live token behind. All 18 auth tests pass, 50 runs in a row with
Go's race checker on.

The fix (service.go (server/internal/auth/service.go), in Refresh): after
saving the new token B, the server checks that the old token A still exists.
Used tokens are kept on purpose, so if A is gone, a reuse cleanup or a logout
ran while this refresh was in progress. In that case the server deletes A's
whole login and returns 401 UNAUTHORIZED. It costs one indexed read per
refresh.

The new test, TestRefreshReuseCleanupBetweenUseAndInsert, makes the race
happen every time instead of hoping for it. A small hook in the fake token
store runs the replay of A just before B is saved. I ran it before writing the
fix and it failed: the winner got a working pair, and B survived the cleanup.
With the fix it passes.

One existing test changed. TestRefreshParallelOnlyOneWins expected exactly one
of 10 parallel refreshes to win. Now zero winners is also correct: if a
loser's cleanup lands before the winner's check, the winner is refused too.
The test now checks that at most one pair is issued and that no refresh token
survives. The second check is the stronger one, and it's the gap that was open
before.

Docs: PLAN.md section 8 has a new step 5 for this check. API.md needs no
change, because the client still sees a 401.

The client lock in phase 10 is still needed: without it, two tabs of the same
user set off this reuse alarm and log out.
