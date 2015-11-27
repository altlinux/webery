angular.module('girar', ['ngRoute', 'ngSanitize','relativeDate','ui.bootstrap','ui.bootstrap.typeahead'])
.config(['$routeProvider', '$locationProvider',
	function($routeProvider, $locationProvider) {
		$routeProvider
			.when('/task/:taskId', {
				templateUrl: '/task.html',
				controller: 'TaskCtrl',
//				controllerAs: 'task'
			})
			.when('/acl/:repo/packages/:name', {
				templateUrl: '/acl-info.html',
				controller: 'AclPackageInfoCtrl',
			})
			.when('/acl/:repo/groups/:name', {
				templateUrl: '/acl-info.html',
				controller: 'AclGroupInfoCtrl',
			})
			.when('/acl', {
				templateUrl: '/acl.html',
				controller: 'AclCtrl',
			})
			.when('/suggestion', {
				controller: 'SuggestionCtrl',
			})
			.when('/main', {
				templateUrl: '/main.html',
				controller: 'SearchCtrl',
			})
			.when('/apidoc', {
				templateUrl: '/apidoc.html'
			})
			.otherwise({
				redirectTo: '/main'
			});
		$locationProvider.html5Mode(true);
	}
])
.filter('convertState', function ($sce) {
	return function(state) {
		var label = "default";
		switch (state) {
			case "eperm":
			case "failed":
				label = "danger";
				break;
			case "tested":
			case "done":
				label = "success";
				break;
			case "new":
			case "awaiting":
			case "postponed":
			case "building":
			case "pending":
			case "commiting":
				label = "info"
				break;
			default:
				return "";
		}
		var out = '<span class="label label-' + label + '">' + state + '</span>';
		return $sce.trustAsHtml(out)
	};
})
.filter('convertShared', function ($sce) {
	return function (value) {
		var out = "";
		if (value === true) {
			out = '<span class="label label-default">shared</span>';
		}
		return $sce.trustAsHtml(out)
	};
})
.filter('convertTestonly', function ($sce) {
	return function (value) {
		var out = "";
		if (value === true) {
			out = '<span class="label label-default">test only</span>';
		}
		return $sce.trustAsHtml(out)
	};
})
.directive('focusItem', function($timeout) {
	return {
		link: function(scope, element, attrs) {
			scope.$watch(attrs.focusItem, function() {
				element[0].focus();
			});
		}
	};
})
.controller('BodyCtrl', ['$scope', '$rootScope', function($scope, $rootScope) {
	$rootScope.GitAltUrl = "//git.altlinux.org";
}])
.controller('MainCtrl', ['$route', '$routeParams', '$location', function($route, $routeParams, $location) {
	this.$route       = $route;
	this.$location    = $location;
	this.$routeParams = $routeParams;
}])
.controller('ApiDocCtrl', ['$scope', function($scope) {
	$scope.oneAtATime = true;
	$scope.status = {};

	$scope.toggle_open = function(i) {
		var name = "group-" + i;
		for (var prop in $scope.status) {
			if (!$scope.status.hasOwnProperty(prop)) {
				continue;
			}
			if (prop === name) {
				continue;
			}
			$scope.status[prop] = false;
		}
		var old = $scope.status[name] || false;
		$scope.status[name] = !old;
	};

	$scope.is_open = function(i) {
		var name = "group-" + i;
		return $scope.status[name] || false;
	};

}])
.controller('SearchCtrl', ['$scope', '$location', '$http', function($scope, $location, $http) {
	$scope.getResults = function(val) {
		return $http.get('/api/v1/search', {
			params: {
				prefix: val,
				limit: 10,
			}
		}).then(function(response) {
			if (!response.data.data) {
				return [];
			}

			var fields = ["taskid", "owner", "pkgname", "repo"];
			var re = new RegExp('^(' + response.data.data.Query + ')');

			return response.data.data.Result.map(function(item) {
				item.task_id = item.taskid;
				item.url = "/task/" + item.taskid;
				item.include = "search-suggestion-" + item.objtype + ".html";


				fields.forEach(function (element, index, array) {
					if (item[element]) {
						item[element] = item[element].toString().replace(re, '<span class="searchmatch">$1</span>');
					}
				});

				return item;
			}).sort(function (a, b) {
				if (a.taskid < b.taskid) {
					return 1;
				}
				if (a.taskid > b.taskid) {
					return -1;
				}
				return 0;
			});
		});
	};
	$scope.showResult = function(item, model, label) {
		$location.url(item.url);
	};
}])
.controller('TaskCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	this.name = "TaskCtrl";
	this.params = $routeParams;

	getTask = function(taskid) {
		return $http.get('/api/v1/tasks/' + taskid, {
			params: {
				nocancelled: true
			}
		}).then(function(response) {
			$scope.task     = response.data.data.task;
			$scope.subtasks = response.data.data.subtasks;

			$scope.subtasks.map(function(item) {
				item.active = (item.status === 'active');
				switch (item.type) {
					case "delete":
					case "copy":
						item.Built = false;
						item.SourceURL = "" +
							$rootScope.GitAltUrl + "/gears/" +
							item.pkgname.charAt(0) + "/..git?p=" +
							item.pkgname + ".git;a=shortlog;h=refs/heads/" +
							$scope.task.repo;
						break;
					default:
						item.Built = true;
						item.SourceURL = "" +
							$rootScope.GitAltUrl + "/tasks/" +
							item.taskid + "/gears/" +
							item.subtaskid + "/git";
				}
			});

			$scope.task.tries = [];
			for (var i = $scope.task.try; i; i--) {
				$scope.task.tries.push(i);
			}
		});
	};

	getTask($routeParams.taskId);
}])
.controller('AclCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	this.name = "AclCtrl";
	this.params = $routeParams;

	$scope.cur_repo = "";
	$scope.repos = [];
	$scope.packages = [];

	$scope.toggleRepo = function(repo) {
		$scope.cur_repo = repo;
	};

	getRepos = function() {
		return $http.get('/api/v1/acl/', {
			params: {}
		}).then(function(response) {
			$scope.repos = response.data.data;
		});
	};

	getPackages = function() {
		if ($scope.cur_repo === "") {
			return;
		}
		return $http.get('/api/v1/acl/' + $scope.cur_repo + '/packages', {
			params: {}
		}).then(function(response) {
			$scope.packages = response.data.data;
		});
	};

	getRepos();
}])
.controller('AclPackageInfoCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	this.name = "AclPackageInfoCtrl";
	this.params = $routeParams;

	$scope.Name     = "";
	$scope.Repo     = "";
	$scope.Members  = [];
	$scope.Found    = false;
	$scope.NotFound = false;

	getACL = function() {
		return $http.get('/api/v1/acl/' + $routeParams.repo + '/packages/' + $routeParams.name, {
			params: {}
		}).then(
			function(response) {
				$scope.Found   = true;
				$scope.Name    = $routeParams.name;
				$scope.Repo    = $routeParams.repo;
				$scope.Members = response.data.data.members;
				$scope.Members.map(function(item) {
					item.include = "acl-" + item.type + ".html";
				});
			},
			function (response) {
				$scope.NotFound = true;
				$scope.Name     = $routeParams.name;
				$scope.Repo     = $routeParams.repo;
			}
		);
	};

	getACL();
}])
.controller('AclGroupInfoCtrl', ['$routeParams', '$scope', '$rootScope', '$http', function($routeParams, $scope, $rootScope, $http) {
	this.name = "AclGroupInfoCtrl";
	this.params = $routeParams;

	$scope.Name     = "";
	$scope.Repo     = "";
	$scope.Members  = [];
	$scope.Found    = false;
	$scope.NotFound = false;

	getACL = function() {
		return $http.get('/api/v1/acl/' + $routeParams.repo + '/groups/' + $routeParams.name, {
			params: {}
		}).then(
			function(response) {
				$scope.Found   = true;
				$scope.Name    = $routeParams.name;
				$scope.Repo    = $routeParams.repo;
				$scope.Members = response.data.data.members;
				$scope.Members.map(function(item) {
					item.include = "acl-" + item.type + ".html";
				});
			},
			function (response) {
				$scope.NotFound = true;
				$scope.Name     = $routeParams.name;
				$scope.Repo     = $routeParams.repo;
			}
		);
	};

	getACL();

	if ($routeParams.name === "everybody" || $routeParams.name === "nobody") {
		$scope.MemberInclude = "acl-info-" + $routeParams.name + ".html";
	}
}])
;
