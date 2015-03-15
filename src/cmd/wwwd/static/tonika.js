
function getRandomInt(min, max) { 
        return Math.floor(Math.random() * (max - min + 1)) + min; 
}
        
function makeLogo(id) {
    var d = $(id);
    var h = "";
    var logo = "Tonika";
    for (var i = 0; i < logo.length; i++) {
      h += "<span class=\"l" + getRandomInt(0,1) + "\">" + logo[i] + "</span>"; }
    d.set('html', h);
}

function addResultOK(data) {
        window.location = '/thanks.html'; 
}

function addResultError(data) {
        window.location = '/thanks.html'; 
}

function onSubmit(event) {
        event.preventDefault();
        e = $.URLEncode($('#email').val());
        $.ajax({
                url: '/beta?e='+e,
                success: addResultOK,
                error: addResultError,
        });
}

$(document).ready(function(){
        $('.wiggle').hover(
                function() {
                        $(this).animate({ left: "-5px" }, 100);
                        $(this).animate({ left: "+5px" }, 100);
                        $(this).animate({ left: "+3px" }, 60);
                        $(this).animate({ left: "-3px" }, 60);
                        $(this).animate({ left: "+1px" }, 20);
                        $(this).animate({ left: "-1px" }, 20);
                        $(this).animate({ left: "0px" }, 10);
                },
                function() {
                        $(this).animate({ left: "0px" }, 100);
                }
        );
        $('.wgg').hover(
                function() {
                        $(this).animate({ top: "+2px" }, 100);
                        $(this).animate({ top: "-2px" }, 100);
                        $(this).animate({ top: "+1px" }, 50);
                        $(this).animate({ top: "-1px" }, 50);
                        $(this).animate({ top: "0px" }, 10);
                },
                function() {
                        $(this).animate({ top: "0px" }, 100);
                }
        );
        $('#ok').click(onSubmit);
});
