package controllers

import (
	db "ass3_part2/db/migrations"
	"ass3_part2/logging"
	"ass3_part2/models"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var subscription models.PremiumSubscription
	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		logging.Logger.Error("Invalid JSON", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Invalid JSON"})
		return
	}
	db.DB.Create(&subscription)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Status: "success", Message: "Subscription created successfully", Data: subscription})
}

func GetSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	var subscription models.PremiumSubscription
	if err := db.DB.First(&subscription, id).Error; err != nil {
		logging.Logger.Error("Subscription not found", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Subscription not found"})
		return
	}
	json.NewEncoder(w).Encode(Response{Status: "success", Data: subscription})
}

func GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var subscriptions []models.PremiumSubscription
	if err := db.DB.Find(&subscriptions).Error; err != nil {
		logging.Logger.Error("Failed to retrieve subscriptions", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Failed to retrieve subscriptions"})
		return
	}
	json.NewEncoder(w).Encode(Response{Status: "success", Data: subscriptions})
}

func UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.URL.Query().Get("id")
	var subscription models.PremiumSubscription
	if err := db.DB.First(&subscription, id).Error; err != nil {
		logging.Logger.Error("Subscription not found", zap.Error(err))
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Subscription not found"})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
		logging.Logger.Error("Invalid JSON", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Invalid JSON"})
		return
	}

	db.DB.Save(&subscription)
	json.NewEncoder(w).Encode(Response{Status: "success", Message: "Subscription updated successfully", Data: subscription})
}

func DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logging.Logger.Error("Invalid subscription ID", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Invalid subscription ID"})
		return
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		logging.Logger.Error("Failed to start transaction", zap.Error(tx.Error))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Failed to start transaction"})
		return
	}

	// Delete related UserSubscriptions
	if err := tx.Where("subscription_id = ?", id).Delete(&models.UserSubscription{}).Error; err != nil {
		tx.Rollback()
		logging.Logger.Error("Failed to delete related user subscriptions", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Failed to delete related user subscriptions"})
		return
	}

	// Delete related Transactions
	if err := tx.Where("subscription_id = ?", id).Delete(&models.Transaction{}).Error; err != nil {
		tx.Rollback()
		logging.Logger.Error("Failed to delete related transactions", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Failed to delete related transactions"})
		return
	}

	// Delete the PremiumSubscription
	if err := tx.Where("id = ?", id).Delete(&models.PremiumSubscription{}).Error; err != nil {
		tx.Rollback()
		logging.Logger.Error("Failed to delete subscription", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Failed to delete subscription"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		logging.Logger.Error("Failed to commit transaction", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{Status: "fail", Message: "Failed to commit transaction"})
		return
	}

	json.NewEncoder(w).Encode(Response{Status: "success", Message: "Subscription and related data deleted successfully"})
}
