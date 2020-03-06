
let curState = null;

function placeBet(sum, color) {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4){
            if (this.status == 200) {
                console.log(this.response);
                resp = JSON.parse(this.response);
                document.getElementById("last_err").innerHTML = resp.err;
            }
        }
    };
    params = "?my_account="+curState.my_account.account+"&sum="+sum+"&color="+color;
    xhttp.open("GET", "/demo/bet"+params, true);
    xhttp.send();
}


function updateState(){
    document.getElementById("my_account").innerHTML = curState.my_account.account;
    document.getElementById("my_balance").innerHTML = curState.my_account.amount;
    document.getElementById("sc_account").innerHTML = curState.sc_account.account;
    document.getElementById("sc_balance").innerHTML = curState.sc_account.amount;
    document.getElementById("num_runs").innerHTML = curState.num_runs;
    document.getElementById("bets_amount").innerHTML = curState.sum_bets;
    document.getElementById("num_bets").innerHTML = curState.num_bets;
    propagateAllAccounts();
    propagateAllBets();

}

function initPage() {
    refreshState();
    refresh(refreshState, 2000);
}

function clearErr() {
    document.getElementById("last_err").innerHTML = "";
}
function refreshState() {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4){
            if (this.status == 200) {
                curState = JSON.parse(this.response);
                updateState();
                clearErr();
            }
        }
    };
    params = "";
    if (curState != null){
        params = "?my_account="+curState.my_account.account;
    }
    xhttp.open("GET", "/demo/state"+params, true);
    xhttp.send();
}

function propagateAllAccounts(){
    allAccountsTable = document.getElementById("all_accounts_table");
    deleteChildren(allAccountsTable);
    for (idx in curState.all_balances){
        row = newAllAccountsRow(idx);
        allAccountsTable.appendChild(row);
    }
}

function newAllAccountsRow(idx){
    account = curState.all_balances[idx].account;
    bal = curState.all_balances[idx].amount;

    row = document.createElement("div");
    row.setAttribute("style", "display: table-row");

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    cell.setAttribute("class", classByAccount(account));
    cell.innerHTML = account;
    row.appendChild(cell);

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    cell.setAttribute("class", classByAccount(account));
    cell.innerHTML = bal;

    row.appendChild(cell);
    return row;
}

function propagateAllBets(){
    allBetsTable = document.getElementById("all_bets_table");
    deleteChildren(allBetsTable);
    allBetsTable.appendChild(newAllBetsHeader());
    for (idx in curState.bets){
        row = newAllBetsRow(idx);
        allBetsTable.appendChild(row);
    }
}

function newAllBetsRow(idx){
    account = curState.bets[idx].p;
    bet = curState.bets[idx].s;
    color = curState.bets[idx].c;

    row = document.createElement("div");
    row.setAttribute("style", "display: table-row");

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    cell.setAttribute("class", classByAccount(account)+" "+classByColor(color));
    cell.innerHTML = account;
    row.appendChild(cell);

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    cell.setAttribute("class", classByAccount(account)+" "+classByColor(color));
    cell.innerHTML = bet;

    row.appendChild(cell);
    return row;
}

function newAllBetsHeader(){
    row = document.createElement("div");
    row.setAttribute("style", "display: table-row");

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    cell.setAttribute("class", "my_account_highlight");
    cell.innerHTML = "Player's account";
    row.appendChild(cell);

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    cell.setAttribute("class", "my_account_highlight");
    cell.innerHTML = "Bet amount";

    row.appendChild(cell);
    return row;
}


function deleteChildren(obj){
    while( obj.hasChildNodes() ){
        obj.removeChild(obj.lastChild);
    }
}

function classByAccount(account){
    if (account == curState.my_account.account){
        return "my_account_highlight";
    } else {
        return"common_highlight";
    }
}

function classByColor(idx){
    if (idx < 0 || idx > 6){
        return "color_7"
    }
    return "color_"+idx.toString()
}

function refresh(fun, millis){
    fun();
    setInterval(fun, millis);
}
