'use strict';

/* Controllers */

function AppCtrl($scope, $q, $timeout) {

  $scope.signalIcons = ['fa-times-circle', 'fa-check-circle', 'fa-smile-o', 'fa-frown-o'];
  $scope.signals = [];
  $scope.join = false;
  $scope.authenticated = false;
  $scope.matched = 0;
  $scope.round = 1;
  $scope.balance = 0;
  $scope.depositaddress = "http://shop.panasonic.com/images/imageNotFound400.jpg";
  $scope.endOfRound = 0;

  $scope.myAction = null;

  var WebSocketHandler = {
  isConnected: false,
  webSocket: null,
  connect: function() {
    var ws = new WebSocket("ws://localhost:8080/play/");
    WebSocketHandler.webSocket = ws;
    ws.onopen = function() {
      console.log("WebSocket opened");
      WebSocketHandler.isConnected = true;
      ws.onclose = function() {
        console.log("WebSocket closed");
        WebSocketHandler.ws = null;
        WebSocketHandler.isConnected = false;
      }
    }
  },
  send: function(dataToSend) {
    // check if connected
    if (WebSocketHandler.isConnected) {
      // send stuff
      WebSocketHandler.webSocket.send(JSON.stringify(dataToSend));

    } else {
      console.log("Error: Not yet connected.");
    }
  },

  listen: function(receiveCallback) {
    WebSocketHandler.webSocket.onmessage = function(message) {
      try {
        receiveCallback(JSON.parse(message.data));
      } catch (err){
        location.load();
      }
    };
  }
};

var maxCount = 30;
$scope.minAmount = 1000;
WebSocketHandler.connect({});

WebSocketHandler.listen(function(data) {
    console.log(data);

    $scope.$apply(function() {
    if(data.command == "signal") {
      signals.push({'player': opponent, 'signal': data.signal});
    };

    if(data.command == "register") {
      if(data.result == 'success') {
        $scope.authenticated = true;
      } else {
        location.reload();
      }
    };

    if(data.command == "endRound") {
      $scope.myAction = null;
      $scope.recentOutcome = data.outcome;
      $scope.recentBalanceDifference = data.balanceDifference;
      $scope.round += 1;
      $scope.initRound();
    };

    if(data.command == 'balance') {
      $scope.balance = data.result;
    };

    if(data.command == 'depositAddress') {
      $scope.depositaddress = data.result;
    }


    if(data.command == "login") {
      if(data.result == 'success') {
        $scope.authenticated = true;
        $scope.getBalance();
      } else {
        location.reload();
      }
    };

    if(data.command == "endGame") {
      $scope.round = 0;
      $scope.myAction = null;
      $scope.matched = 0;
      $scope.round = 0;
      $scope.join = false;
      $scope.endGame = 1;
    };

    if(data.command == "matched") {
      console.log("We match as motherfuckers");
      $scope.matched = 1;
      $scope.counter = minAmount;
    };
  });

});

$scope.$watch('myAction', function(value) {
  if(value != null && $scope.endOfRound != false) {
    $scope.waitForOpponent = true; 
    console.log("WACHTEN OP TEGENSPELER")
  } else {
    $scope.waitForOpponent = false;
  }

});


$(function () {
    $('#info').popover({'html': true});

});

$(function () {
    $('#manage_balance').popover({'html': true});

});

// For the timer
$scope.roundProgressData = {
      label: 0,
      percentage: 0
}

$scope.initRound = function() {
    $scope.getBalance();
    $scope.endOfRound = false;
    $scope.counter = maxCount;
    $scope.myAction = null;
    $scope.signals = []
}


$scope.getBalance = function() {
  WebSocketHandler.send({'command': 'getBalance'});
  WebSocketHandler.send({"command" : "getDepositAddress"});
};


$scope.login = function(name, password) {
  WebSocketHandler.send({command: 'login', name: name, password: password});
};

$scope.logout = function() {
  location.reload();
}

$scope.register = function(name, password) {
  WebSocketHandler.send({command: 'register', name: name, password: password});
};

// Player indicates he wants to start a new game
$scope.joinGame = function() {

  if($scope.balance < $scope.minAmount) {
    return;
  }

  WebSocketHandler.send({'command': 'join'});
  $scope.join = true;
}

$scope.performAction = function(action) {
  if($scope.myAction != null) {
    return;
  }
 $scope.myAction = action;
  WebSocketHandler.send({'command': 'action', 'action': action});
}

$scope.sendRequestOnOpen = function(request) {
  return sendRequest(request);
}

$scope.$watch('roundProgressData', function (newValue, oldValue) {
  newValue.percentage = newValue.label / 100;
}, true);

// The counter
$scope.counter = maxCount;

$scope.onTimeout = function(){
    if($scope.counter > 0) {
      $scope.counter--;
    }
    mytimeout = $timeout($scope.onTimeout,1000);
}

var mytimeout = $timeout($scope.onTimeout,1000);

$scope.stop = function() {
   $timeout.cancel(mytimeout);
}



/*
// The user gets a confirmation/error on withdraw
socket.on('withdraw', function (data) {
  $scope.withdraw = data.result;
});

// Other player played
socket.on('outcome', function (data) {
  socket.emit('get:balance')
  $scope.outcome = data.result;
});

// Methods published to the scope
// ==============================

// Request a depost address
$scope.getDepositAddress = function() {
  socket.emit('get:deposit:address')
}
}

*/

$scope.sendSignal = function(signal) {
  $scope.signals.push({'player': 'you', 'signal': signal});
  WebSocketHandler.send({command: 'signal', signal: signal}, function(data) {
    console.log(data);
  });
}

$scope.$watchCollection('signals', function() {
  console.log($('div.signaloverview').scrolltop);
  $(".signaloverview").animate({ scrollTop: $('.signaloverview').height()}, 1000);
});

}

