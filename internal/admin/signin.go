package admin

import (
	"net/http"
	"regexp"
	"slices"

	"github.com/epistemer/auth"
	"github.com/epistemer/i18n"
	"github.com/epistemer/ui/forms"
	"github.com/epistemer/ui/forms/rule"

	"github.com/epistemer/example/internal/admin/adminroutes"
)

var (
	L10nEmail      = i18n.Key("page.sign_in.email", "Email address")
	L10nPassword   = i18n.Key("page.sign_in.password", "Password")
	L10nRemember   = i18n.Key("page.sign_in.remember", "Remember this device")
	L10nCode       = i18n.Key("page.sign_in.code", "Code")
	L10nCodeHelp   = i18n.Key("page.sign_in.code_help", "We emailed you a code of six digits.")
	L10nSignInBtn  = i18n.Key("page.sign_in.submit", "Sign in")
	L10nVerifyBtn  = i18n.Key("page.sign_in.verify", "Verify")
	L10nDeviceName = i18n.Key("page.sign_in.device", "Browser signed in to Confab")
)

// SignInForm is the first step: who, and the password.
type SignInForm struct {
	*forms.Form
	Email    *forms.Field[string]
	Password *forms.Field[string]
}

func NewSignInForm() *SignInForm {
	f := &SignInForm{Form: forms.New()}
	f.Email = f.Text("email").Label(L10nEmail).Required().Is(rule.Email())
	f.Password = f.Text("password").Raw().Hint(forms.KindPassword).Label(L10nPassword).Required()
	return f
}

// CodeForm is the second: the emailed code, and whether to trust the
// device from now on.
type CodeForm struct {
	*forms.Form
	Code     *forms.Field[string]
	Remember *forms.Field[bool]
}

var sixDigits = regexp.MustCompile(`[0-9]{6}`)

func NewCodeForm() *CodeForm {
	f := &CodeForm{Form: forms.New()}
	f.Code = f.Text("code").Label(L10nCode).Help(L10nCodeHelp).Required().Is(rule.Matches(sixDigits))
	f.Remember = f.Checkbox("remember").Label(L10nRemember)
	return f
}

//h:key sign_in
//h:method GET
//h:path /sign-in
//h:auth public
//h:desc Sign in
func (h *H) SignIn(w http.ResponseWriter, r *http.Request) error {
	return render(w, r, http.StatusOK, signInPage(r, NewSignInForm()))
}

//h:key sign_in.password
//h:method POST
//h:path /sign-in
//h:auth public
//h:desc Check a password
func (h *H) SignInPassword(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	f := NewSignInForm()
	if !f.Submit(ctx, r) {
		return render(w, r, http.StatusUnprocessableEntity, signInPage(r, f))
	}
	flow, err := h.Auth.SignIn(ctx, f.Email.Value(), auth.Cookie())
	if err == nil {
		flow, err = h.Auth.Prove(ctx, flow.Token, auth.Password(f.Password.Value()))
	}
	if err != nil {
		// A wrong password and an unknown address answer alike.
		f.Fail(err)
		return render(w, r, handlersStatus(err), signInPage(r, f))
	}
	if !flow.Verified && slices.Contains(flow.Offers, auth.MethodEmailCode) {
		if _, err := h.Auth.Challenge(ctx, flow.Token, auth.MethodEmailCode); err != nil {
			f.Fail(err)
			return render(w, r, handlersStatus(err), signInPage(r, f))
		}
		flow.Write(w)
		http.Redirect(w, r, adminroutes.SignInCode().URL(), http.StatusSeeOther)
		return nil
	}
	return h.finish(w, r, flow.Token, false)
}

//h:key sign_in.code
//h:method GET
//h:path /sign-in/code
//h:auth public
//h:desc Enter the emailed code
func (h *H) SignInCode(w http.ResponseWriter, r *http.Request) error {
	if _, ok := auth.FlowToken(r.Context()); !ok {
		http.Redirect(w, r, adminroutes.SignIn().URL(), http.StatusSeeOther)
		return nil
	}
	return render(w, r, http.StatusOK, codePage(r, NewCodeForm()))
}

//h:key sign_in.verify
//h:method POST
//h:path /sign-in/code
//h:auth public
//h:desc Check the emailed code
func (h *H) SignInVerify(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	token, ok := auth.FlowToken(ctx)
	if !ok {
		return auth.ErrFlowExpired
	}
	f := NewCodeForm()
	if !f.Submit(ctx, r) {
		return render(w, r, http.StatusUnprocessableEntity, codePage(r, f))
	}
	flow, err := h.Auth.Prove(ctx, token, auth.EmailCode(f.Code.Value()))
	if err != nil {
		f.Code.Fail(err)
		return render(w, r, handlersStatus(err), codePage(r, f))
	}
	return h.finish(w, r, flow.Token, f.Remember.Value())
}

func (h *H) finish(w http.ResponseWriter, r *http.Request, token string, remember bool) error {
	var opts []auth.FinishOption
	if remember {
		opts = append(opts, auth.RememberDevice(L10nDeviceName.T(r.Context())))
	}
	res, err := h.Auth.Finish(r.Context(), token, opts...)
	if err != nil {
		return err
	}
	res.Write(w)
	http.Redirect(w, r, adminroutes.Dashboard().URL(), http.StatusSeeOther)
	return nil
}

//h:key sign_out
//h:method POST
//h:path /sign-out
//h:auth session
//h:desc Sign out
func (h *H) SignOut(w http.ResponseWriter, r *http.Request) error {
	if err := h.Auth.SignOut(r.Context(), w); err != nil {
		return err
	}
	http.Redirect(w, r, adminroutes.SignIn().URL(), http.StatusSeeOther)
	return nil
}
