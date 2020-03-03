
let account = "";
let balance = 0;

function refreshAccountValues(){
    document.getElementById("account").innerHTML = account;
    document.getElementById("balance").innerHTML = balance;
}
function refreshAccount() {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4){
            if (this.status == 200) {
                accountInfo = JSON.parse(this.response);
                account = accountInfo["account"];
                balance = accountInfo["amount"];
                refreshAccountValues();
            }
        }
    };
    xhttp.open("GET", "/poc/newaccount", true);
    xhttp.send();
}
