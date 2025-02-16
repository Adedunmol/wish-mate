package wishlist

import (
	"errors"
	"github.com/Adedunmol/wish-mate/internal/auth"
	"github.com/Adedunmol/wish-mate/internal/helpers"
	"github.com/Adedunmol/wish-mate/internal/reminder"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type Handler struct {
	Store     Store
	UserStore auth.Store
	TaskStore reminder.TaskStore
}

func (h *Handler) CreateWishlist(responseWriter http.ResponseWriter, request *http.Request) {
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

	userData, err := h.UserStore.FindUserByEmail(email.(string))
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	if body.Date == "" {
		body.Date = userData.DateOfBirth
	}

	wishlist := Wishlist{
		Name:         body.Name,
		Description:  body.Description,
		Items:        body.Items,
		NotifyBefore: body.NotifyBefore,
		Date:         body.Date,
	}

	data, err := h.Store.CreateWishlist(userData.ID, wishlist)

	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	// create a scheduleDate using the calculated days before the birthday/due date and create a notification and send mails
	_, err = helpers.CalculateDaysBefore(wishlist.Date, wishlist.NotifyBefore)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	// create a task in the reminder table
	//taskBody := reminder.CreateTaskBody{
	//	Name:      "notification",
	//	UserID:    userData.ID,
	//	Title:     "Wishlist",
	//	Body:      fmt.Sprintf("%s created a wishlist for a special date %d. kindly check it out.", userData.Username, wishlist.Date),
	//	Type:      "normal",
	//	ExecuteAt: &scheduledDate,
	//}
	//
	//_, err = h.TaskStore.CreateTask(taskBody)
	//if err != nil {
	//	helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
	//	return
	//}

	response := Response{
		Status:  "Success",
		Message: "Wishlist created successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusCreated)
}

func (h *Handler) GetAllWishlists(responseWriter http.ResponseWriter, request *http.Request) {
	// should be verbose (include the details of those who picked an item) if due date >= current date
	// should not include items that have been picked for other users but should include for the creator

	userID := chi.URLParam(request, "user_id")

	if userID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("id is required"), http.StatusBadRequest, "user id is required", nil))
		return
	}

	currentUserID := request.Context().Value("user_id")

	if currentUserID == nil || currentUserID == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	newUserID, err := strconv.Atoi(userID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newCurrentUserID := currentUserID.(int)

	_, err = h.UserStore.FindUserByID(newUserID)

	if err != nil {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(err, http.StatusNotFound, "no user found with the id", nil))
		return
	}

	wishlists, err := h.Store.GetUserWishlists(newUserID, newCurrentUserID == newUserID)

	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrNotFound)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Wishlists retrieved successfully",
		Data:    wishlists,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) GetWishlist(responseWriter http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	if id == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("id is required"), http.StatusBadRequest, "id is required", nil))
		return
	}

	userID := request.Context().Value("user_id")

	if userID == nil || userID == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	wishlistID, err := strconv.Atoi(id)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newUserID := userID.(int)

	// should be verbose (include the details of those who picked an item) if due date >= current date
	// should not include items that have been picked for other users but should include for the creator
	wishlist, err := h.Store.GetWishlistByID(wishlistID, newUserID)

	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrNotFound)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Wishlist retrieved successfully",
		Data:    wishlist,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) UpdateWishlist(responseWriter http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	if id == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("id is required"), http.StatusBadRequest, "id is required", nil))
		return
	}

	body, problems, err := helpers.DecodeAndValidate[*UpdateWishlist](request)

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

	userID := request.Context().Value("user_id")

	if userID == nil || userID == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	wishlistID, err := strconv.Atoi(id)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newUserID := userID.(int)

	wishlist, err := h.Store.UpdateWishlistByID(wishlistID, newUserID, *body)

	if err != nil && errors.Is(err, helpers.ErrForbidden) {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	if err != nil && errors.Is(err, helpers.ErrNotFound) {
		helpers.HandleError(responseWriter, helpers.ErrNotFound)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Wishlist updated successfully",
		Data:    wishlist,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) DeleteWishlist(responseWriter http.ResponseWriter, request *http.Request) {
	id := chi.URLParam(request, "id")

	if id == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("id is required"), http.StatusBadRequest, "id is required", nil))
		return
	}

	userID := request.Context().Value("user_id")

	if userID == nil || userID == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	wishlistID, err := strconv.Atoi(id)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newUserID := userID.(int)

	err = h.Store.DeleteWishlistByID(wishlistID, newUserID)

	if err != nil && errors.Is(err, helpers.ErrForbidden) {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	if err != nil && errors.Is(err, helpers.ErrNotFound) {
		helpers.HandleError(responseWriter, helpers.ErrNotFound)
		return
	}

	// delete scheduled job for the wishlist
	//err = h.TaskStore.DeleteTask(wishlistID)
	//
	//if err != nil {
	//	helpers.HandleError(responseWriter, err)
	//	return
	//}

	response := Response{
		Status:  "Success",
		Message: "Wishlist deleted successfully",
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) GetItemHandler(responseWriter http.ResponseWriter, request *http.Request) {
	wishlistID := chi.URLParam(request, "wishlist_id")

	if wishlistID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("wishlist id is required"), http.StatusBadRequest, "wishlist id is required", nil))
		return
	}

	itemID := chi.URLParam(request, "item_id")

	if itemID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("item id is required"), http.StatusBadRequest, "item id is required", nil))
		return
	}

	newWishlistID, err := strconv.Atoi(wishlistID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newItemID, err := strconv.Atoi(itemID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	data, err := h.Store.GetItem(newWishlistID, newItemID)

	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Item retrieved successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) DeleteItemHandler(responseWriter http.ResponseWriter, request *http.Request) {

	wishlistID := chi.URLParam(request, "wishlist_id")

	if wishlistID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("wishlist id is required"), http.StatusBadRequest, "wishlist id is required", nil))
		return
	}

	itemID := chi.URLParam(request, "item_id")

	if itemID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("item id is required"), http.StatusBadRequest, "item id is required", nil))
		return
	}

	userID := request.Context().Value("user_id")

	if userID == nil || userID == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	newWishlistID, err := strconv.Atoi(wishlistID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newItemID, err := strconv.Atoi(itemID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newUserID := userID.(int)

	wishlist, err := h.Store.GetWishlistByID(newWishlistID, newUserID)

	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	if wishlist.UserID != newUserID {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	err = h.Store.DeleteItem(newWishlistID, newItemID)
	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Item deleted successfully",
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) UpdateWishlistItemHandler(responseWriter http.ResponseWriter, request *http.Request) {
	// for the users that created the wishlist
	wishlistID := chi.URLParam(request, "wishlist_id")

	if wishlistID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("wishlist id is required"), http.StatusBadRequest, "wishlist id is required", nil))
		return
	}

	itemID := chi.URLParam(request, "item_id")

	if itemID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("item id is required"), http.StatusBadRequest, "item id is required", nil))
		return
	}

	body, problems, err := helpers.DecodeAndValidate[*UpdateItem](request)

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

	userID := request.Context().Value("user_id")

	if userID == nil || userID == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	newWishlistID, err := strconv.Atoi(wishlistID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newItemID, err := strconv.Atoi(itemID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newUserID := userID.(int)

	wishlist, err := h.Store.GetWishlistByID(newWishlistID, newUserID)

	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	if wishlist.UserID != newUserID {
		helpers.HandleError(responseWriter, helpers.ErrForbidden)
		return
	}

	data, err := h.Store.UpdateItem(newWishlistID, newItemID, body)
	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Item updated successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}

func (h *Handler) PickWishlistItemHandler(responseWriter http.ResponseWriter, request *http.Request) {
	// for the friends that are picking the items
	wishlistID := chi.URLParam(request, "wishlist_id")

	if wishlistID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("wishlist id is required"), http.StatusBadRequest, "wishlist id is required", nil))
		return
	}

	itemID := chi.URLParam(request, "item_id")

	if itemID == "" {
		helpers.HandleError(responseWriter, helpers.NewHTTPError(errors.New("item id is required"), http.StatusBadRequest, "item id is required", nil))
		return
	}

	userID := request.Context().Value("user_id")

	if userID == nil || userID == "" {
		helpers.HandleError(responseWriter, helpers.ErrUnauthorized)
		return
	}

	newWishlistID, err := strconv.Atoi(wishlistID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newItemID, err := strconv.Atoi(itemID)
	if err != nil {
		helpers.HandleError(responseWriter, helpers.ErrInternalServerError)
		return
	}

	newUserID := userID.(int)

	data, err := h.Store.PickItem(newWishlistID, newItemID, newUserID)

	if err != nil {
		helpers.HandleError(responseWriter, err)
		return
	}

	response := Response{
		Status:  "Success",
		Message: "Item picked successfully",
		Data:    data,
	}

	helpers.WriteJSONResponse(responseWriter, response, http.StatusOK)
}
