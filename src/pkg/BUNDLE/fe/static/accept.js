function acceptResultOK(data) {
        $('#wait').hide();

        if (data.ErrMsg != "") {
                $('#f_err').text(data.ErrMsg);
                $('#notice').show();
        }
        
        if (data.InviteMsg != "") {
                $('#f_invite').val(data.InviteMsg);
                $('#invite').show();
        }

        $('#done').show();
        $('#f_done').click(function(event){ 
                event.preventDefault();
                window.location = AdminURL; 
        });

        if (data.ErrMsg != "") {
                $.scrollTo('#hm',{duration:"400"});
        }
        if (data.InviteMsg != "") {
                $.scrollTo('#great',{duration:"400"});
        }
}

function acceptResultError(data) {
        $('#wait').hide();
        $('#error').show();
}

function onYes(event) {
        event.preventDefault();
        $('#f_name').attr("disabled", true);
        $('#f_email').attr("disabled", true);
        $('#f_addr').attr("disabled", true);
        $('#f_yes').attr("disabled", true);
        $('#f_yes').addClass("disabled");
        $('#f_yes').val("Accepting...");
        $('#f_yes').blur();
        $('#f_no').hide();

        sl = $.URLEncode($('#f_slot').val());
        dk = $.URLEncode($('#f_dialkey').val());
        sk = $.URLEncode($('#f_sigkey').val());
        na = $.URLEncode($('#f_name').val());
        em = $.URLEncode($('#f_email').val());
        ad = $.URLEncode($('#f_addr').val());
        $('#wait').show();
        $.ajax({
                url: '/api/accept?na='+na+'&em='+em+'&ad='+ad+'&sl='+sl+'&dk='+dk+'&sk='+sk,
                success: acceptResultOK,
                error: acceptResultError,
                dataType: 'json',
        });
}

$(document).ready(function(){
        mainReady();
        $('#f_name').removeAttr("disabled");
        $('#f_email').removeAttr("disabled");
        $('#f_addr').removeAttr("disabled");

        $('#f_yes').val("Accept!");
        $('#f_yes').removeAttr("disabled");
        $('#f_yes').click(onYes);

        $('#f_no').click(function(){ window.location = AdminURL; });
        $('#f_no').removeAttr("disabled");
        $('#f_no').show();
});
