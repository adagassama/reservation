package web

import (
	"html/template"
	"net/http"
)

type TemplateData struct {
	Titre   string
	Content any
}

func (h Handler) WebShowShops() http.HandlerFunc {
	data := TemplateData{Titre: "Tous les Shops"}

	return func(writer http.ResponseWriter, request *http.Request) {
		/*shops, err := h.Store.GetShops()
		data.Content = shops*/

		tmpl, err := template.ParseFiles("templates/layout.gohtml", "templates/list.gohtml")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		err = tmpl.ExecuteTemplate(writer, "layout", data)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) WebCreateShopForm() http.HandlerFunc {
	data := TemplateData{Titre: "Add a shop"}

	return func(writer http.ResponseWriter, request *http.Request) {
		tmpl, err := template.ParseFiles("templates/layout.gohtml", "templates/createShop.gohtml")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		err = tmpl.ExecuteTemplate(writer, "layout", data)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (h *Handler) WebCreateConnexionForm() http.HandlerFunc {
	data := TemplateData{Titre: "Add a connexion"}

	return func(writer http.ResponseWriter, request *http.Request) {
		tmpl, err := template.ParseFiles("templates/layout.gohtml", "templates/connexion.gohtml")
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

		err = tmpl.ExecuteTemplate(writer, "layout", data)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}
