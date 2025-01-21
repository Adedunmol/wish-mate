package wishlist

import (
	"errors"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/user"
	"net/http"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
	Store     Store
	UserStore user.Store
}

func (h Handler) CreateWishlist(responseWriter http.ResponseWriter, request *http.Request) {
	body, problems, err := helpers.DecodeAndValidate[*Wishlist](request)

	var clientError helpers.ClientError
	ok := errors.As(err, &clientError)

	if err != nil && problems == nil {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusBadRequest, "invalid request body", nil))
		return
	}

	if err != nil && ok {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusBadRequest, "invalid request body", problems))
		return
	}

	email := request.Context().Value("email")

	if email == nil || email == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	user, err := h.UserStore.FindUserByEmail(email.(string))
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	wishlist := Wishlist{
		Name:        body.Name,
		Description: body.Description,
		Items:       body.Items,
	}

	data, err := h.Store.CreateWishlist(user.ID, wishlist)

	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "User created successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusCreated)
}
