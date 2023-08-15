$.ajaxSetup({
    headers: {
        'Content-Type': "application/json",
        'Authorization': `Bearer ${localStorage.getItem("token")}`,
    }
});

$('#sign-out').click(function () {
    signOut()
})

function signOut() {
    localStorage.removeItem("token")
    window.location = "index.html"
}

function checkAuth(response) {
    if (response.status === 401) {
        signOut()
    }
}

function parseBool(s) {
    if(typeof s == "string") {
        return Boolean(s.replace(/(false)|(off)|(no)|(n)|(0)/i, ""))
    } else {
        return Boolean(s)
    }
}