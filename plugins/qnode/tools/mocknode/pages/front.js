
let curState = null;

function placeBet(sum, color) {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4){
            if (this.status == 200) {
                refreshState();
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
    propagateAllAccounts();
    propagateAllBets();
}

function refreshState() {
    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4){
            if (this.status == 200) {
                curState = JSON.parse(this.response);
                updateState();
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
    if (account == curState.my_account.account){
        cell.setAttribute("class", "my_account_highlight");
    } else {
        cell.setAttribute("class", "common_highlight");
    }
    cell.innerHTML = account;
    row.appendChild(cell);

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    if (account == curState.my_account.account){
        cell.setAttribute("class", "my_account_highlight");
    } else {
        cell.setAttribute("class", "common_highlight");
    }
    cell.innerHTML = bal;

    row.appendChild(cell);
    return row;
}

function colorIdxToStyle(idx){
    switch (idx) {
        case 0:
            return "common_highlight";  // "red_line";
        case 1:
            return "common_highlight";  // "yellow_line";
        case 2:
            return "common_highlight";  // "green_line";
        case 3:
            return "common_highlight";  // "blue_line";
        case 4:
            return "common_highlight";  // "magenta_line";
        case 5:
            return "common_highlight";  // "orange_line";
        case 6:
            return "common_highlight";  // "cyan_line";
        case 7:
            return "common_highlight";  // "brown_line";
    }
    return "black_line";
}

function propagateAllBets(){
    allBetsTable = document.getElementById("all_bets_table");
    deleteChildren(allBetsTable);
    for (idx in curState.bets){
        row = newAllBetsRow(idx);
        allBetsTable.appendChild(row);
    }
}

function newAllBetsRow(idx){
    account = curState.bets[idx].a;
    bet = curState.bets[idx].v;
    color = curState.bets[idx].color

    row = document.createElement("div");
    row.setAttribute("style", "display: table-row");

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    cell.setAttribute("class", colorIdxToStyle(color));
    cell.innerHTML = account;
    row.appendChild(cell);

    cell = document.createElement("div");
    cell.setAttribute("style", "display: table-cell");
    cell.setAttribute("class", colorIdxToStyle(color));
    cell.innerHTML = bet;

    row.appendChild(cell);
    return row;
}

function deleteChildren(obj){
    while( obj.hasChildNodes() ){
        obj.removeChild(obj.lastChild);
    }
}
