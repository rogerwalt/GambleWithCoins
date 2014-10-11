'use strict';

/* Controllers */

function AppCtrl($scope) {

  // Socket listeners
  // ================

  // For the timer
  $scope.roundProgressData = {
        label: 0,
        percentage: 0
  }

  var connection = new WebSocket("ws://localhost:8080/play/",'json');

  var sendRequest = function(request) {
    connection.onopen = function () {
      connection.send(JSON.stringify(request)); //send a message to server once connection is opened.
    };

    connection.onerror = function (error) {
      console.log('Error Logged: ' + error); //log errors
      return false;
    };

    connection.onmessage = function (e) {
      console.log('Received From Server: ' + e.data); //log the received message
      return e.data;
    };
  }

  $scope.getBalance = function() {
    var request = {command: 'getBalance'};
    return sendRequest(request);
  }

  console.log('Check that balance');
  console.log($scope.getBalance());

  //$scope.customers = WebSocketFactory.getCustomers();

  $scope.signalIcons = ['fa-times-circle', 'fa-check-circle', 'fa-smile-o', 'fa-frown-o']

  $scope.$watch('roundProgressData', function (newValue, oldValue) {
    newValue.percentage = newValue.label / 100;
  }, true);

  $scope.signals = [{player: 'you', signal: 1}, {player:'opposite', signal: 2}];
  $scope.join = false;

  /*
  socket.on('')

  // Initiate the game (set balance for user)
  socket.on('init', function (data) {
    socket.emit('get:balance')
  });

  // The user gets the requested balance of the server
  socket.on('balance', function (data) {
    $scope.balance = data.result;
  });

  // The user gets the requested depositaddress of the server
  socket.on('deposit:address', function (data) {
    $scope.depositAddress = data.result;
  });

  // The user recieves a message from the other user
  socket.on('send:signal', function (data) {
    $scope.signals.push({
      player: 'oponent',
      signal: data.signal
    });
  });

  // Let the games begin
  socket.on('matched', function (data) {
    $scope.matched = true;
  });

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

  // Player indicates he wants to start a new game
  $scope.joinGame = function() {
    console.log('joingame')
    socket.emit('join');
    $scope.join = true;
    console.log($scope.join)
  }

  // Request a depost address
  $scope.getDepositAddress = function() {
    socket.emit('get:deposit:address')
  }

  // The player has made a decision 
  $scope.sendAction = function(action) {
    socket.emit('command:action', {
       action: action
    });

    console.log('Action: ' + action)
    console.log('Join: ' + $scope.join)
    console.log('Matched: ' + $scope.matched)
  };

  // The player sends a signal to the other player
  $scope.sendSignal = function(signal) {
    socket.emit('send:signal', {
      signal: signal
    })

    $scope.signals.push({
      player: 'you',
      signal: signal
    });
  }

  $scope.countDown = function(seconds) {
    $scope.roundProgressData.percentage = 0
    $scope.roundProgressData.label = seconds
    for (var i = 1; i <= seconds; i++) {
      $scope.roundProgressData.percentage += 100.0/seconds ;
      $scope.roundProgressData.label = seconds - i;

      console.log($scope.roundProgressData)
    }
  }
  
  */

  $scope.$watchCollection('signals', function() {
    console.log($('div.signaloverview').scrolltop);
    $(".signaloverview").animate({ scrollTop: $('.signaloverview').height()}, 1000);
  });

  /*$scope.changeName = function () {
    socket.emit('change:name', {
      name: $scope.newName
    }, function (result) {
      if (!result) {
        alert('There was an error changing your name');
      } else {
        
        changeName($scope.name, $scope.newName);

        $scope.name = $scope.newName;
        $scope.newName = '';
      }
    });
  };
  */

}
