// Конфигурация на приложението
const config = {
    apiUrl: 'http://localhost:8080',
    endpoints: {
        register: '/register',
        login: '/login',
        resetPassword: '/reset-password',
        weight: '/weight',
        weightStats: '/weight/stats',
        weightDelete: '/weight/:id',
        userSettings: '/user/settings',
        changePassword: '/user/password',
        users: '/users',
        visibility: '/user/visibility',
        friends: '/friends',
        friendRequest: '/friends/request/:userId',
        friendAccept: '/friends/accept/:requestId',
        friendReject: '/friends/reject/:requestId',
        challenges: '/challenges',
        challengeAccept: '/challenges/:challengeId/accept',
        challengeReject: '/challenges/:challengeId/reject',
        challengeResults: '/challenges/:challengeId/results',
        updateVisibility: '/user/visibility'
    }
}; 