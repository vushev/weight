// Конфигурация на приложението

const config = {
    apiUrl: '',  // Празен string, защото сме на същия домейн
    endpoints: {
        // Автентикация
        register: '/register',
        login: '/login',
        resetPassword: '/reset-password',
        authStatus: '/auth-status',
        logout: '/logout',
        // Тегло
        weight: '/api/weight',
        weightStats: '/api/weight/stats',
        weightDelete: '/api/weight/:id',
        
        // Настройки на потребителя
        userSettings: '/api/settings',
        updateUserSettings: '/api/settings',
        updateVisibility: '/api/settings/visibility',
        changePassword: '/api/user/password',
        
        // Потребители и приятели
        users: '/api/users',
        friends: '/api/friends',
        friendRequest: '/api/friends/request/:userId',
        friendAccept: '/api/friends/accept/:friendshipId',
        friendReject: '/api/friends/reject/:friendshipId',
        
        // Предизвикателства
        challenges: '/api/challenges',
        createChallenge: '/api/challenges',
        challengeAccept: '/api/challenges/:challengeId/accept',
        challengeReject: '/api/challenges/:challengeId/reject',
        challengeResults: '/api/challenges/:challengeId/results',
        
        // Калории - настройки и изчисления
        calorieSettings: '/api/calories/settings',
        updateCalorieSettings: '/api/calories/settings',
        calorieCalculations: '/api/calories/calculations',
        
        // Калории - дневен лог
        dailyCalorieLog: '/api/calories/log/:date',
        
        // Калории - храна и активности
        calorieIntake: '/api/calories/log',
        calorieIntakeDelete: '/api/calories/food/:id',
        calorieActivity: '/api/calories/activity',
        calorieActivityDelete: '/api/calories/activity/:id',
        addFoodEntry: '/api/calories/food',
        deleteFoodEntry: '/api/calories/food/:id',
        addActivityEntry: '/api/calories/activity',
        deleteActivityEntry: '/api/calories/activity/:id',
        mealTypes: '/api/calories/meal-types',
        // Калории - статистика
        calorieStats: '/api/calories/stats'
    }
}; 
