/*
 * Grid Layout 2.0.0
 *
 * Copyright (c) 2007 Stephen Hallgren (teevio.com)
 * 
 * Modification for fluid layouts by J.Bradford Dillon (jbradforddillon.com)
 * 
 */

$(document).ready(function()
{
	gridLayout = new GridLayout();
});

function GridLayout()
{
	gridLayoutObj      = this;
	this.grid_settings = eval("(" + $('#GridLayout-params').html() + ")");
	
	this.calculateTotalWidth();
	this.calculateSubColumnWidth();
	this.attachEvents();
	this.buildHTML();
	this.gridHeight();
}

GridLayout.prototype.attachEvents = function ()
{
	$(window).scroll(function(){ gridLayoutObj.gridHeight(); });
	$(window).resize(function(){ gridLayoutObj.gridHeight(); });
	$(window).keypress(function(event)
	{			
		gridLayoutObj.checkKeyPress(event);
	});
	
	$('body').keypress(function(event)
	{		
		if (window.event && document.all)
		{
		    e = window.event;

			if(e.keyCode) // IE
        	{
                gridLayoutObj.checkKeyPress(e);
        	}
        }		
	});
};

GridLayout.prototype.buildHTML = function ()
{
	htmlOutput = '';
		
	if (this.grid_settings.column_count > 0)
	{
		htmlOutput = '<ul>';

		for (var i = 1; i <= this.grid_settings.column_count; i++)
		{
			htmlOutput += '<li>&nbsp;';
			
			if (this.grid_settings.subcolumn_count > 0)
			{
				htmlOutput += '<ul>';

				for (var j = 1; j <= this.grid_settings.subcolumn_count; j++)
				{
					htmlOutput += '<li>&nbsp;</li>';
				}
				
				htmlOutput += '</ul>';	
			}
			
			htmlOutput += '</li>';
		}
		
		htmlOutput += '</ul>';
	}
	
	$('#GridLayout-params').html(htmlOutput);
	
	this.applyCSS();
};

GridLayout.prototype.applyCSS = function ()
{
	// #JBD
	if(gridLayoutObj.grid_settings.percent == true)
	{
		// Test for percentage setting
		// This is a duplication of the px version, with math for percent where applicable. A more efficient method could be built, but this was quick and dirty.
		// We're using percentages
		$('#GridLayout *').css({margin:0,padding:0});
		$('#GridLayout').css({display:'none',position:'absolute',"z-index":'1000',top:'0px',width:this.grid_settings.width+'%',minWidth:this.grid_settings.min_width+'px'});
		$('#GridLayout ul').css({'list-style':'none'});
		$('#GridLayout li').css({'background-image':'url(/static/gridlayout.png)','background-repeat':'repeat', display:'block','float':'left',position:'relative'});
		$('#GridLayout ul li ul').css({position:'absolute',top:0,left:0});
			
		$('#GridLayout ul > li').each(function(i)
		{
			if (i % gridLayoutObj.grid_settings.column_count < (gridLayoutObj.grid_settings.column_count-1))
			{
				$(this).css({width:gridLayoutObj.grid_settings.column_width+'%',"margin-right":gridLayoutObj.grid_settings.column_gutter+'%'});
			}
			else
			{
				$(this).css({width:gridLayoutObj.grid_settings.column_width+'%',"margin-right":'0%'});			
			}
		});
		
		$('#GridLayout ul li ul li').each(function(i)
		{
			if (i % gridLayoutObj.grid_settings.subcolumn_count < (gridLayoutObj.grid_settings.subcolumn_count-1))
			{
				$(this).css({width:gridLayoutObj.grid_settings.subcolumn_width+'%',"margin-right":gridLayoutObj.grid_settings.column_gutter+'%'});
			}
			else
			{
				$(this).css({width:gridLayoutObj.grid_settings.subcolumn_width+'%',"margin-right":'0%'});			
			}
		});
	}
	else
	{
		// We're not using percentages. Default to pixels.
		$('#GridLayout *').css({margin:0,padding:0});
		$('#GridLayout').css({display:'none',position:'absolute',"z-index":'1000',top:'0px',width:this.grid_settings.total_width+'px'});
		$('#GridLayout ul').css({'list-style':'none'});
		$('#GridLayout li').css({'background-image':'url(/static/gridlayout.png)','background-repeat':'repeat', display:'block','float':'left',position:'relative'});
		$('#GridLayout ul li ul').css({position:'absolute',top:0,left:0});
			
		$('#GridLayout ul > li').each(function(i)
		{
			if (i % gridLayoutObj.grid_settings.column_count < (gridLayoutObj.grid_settings.column_count-1))
			{
				$(this).css({width:gridLayoutObj.grid_settings.column_width+'px',"margin-right":gridLayoutObj.grid_settings.column_gutter+'px'});
			}
			else
			{
				$(this).css({width:gridLayoutObj.grid_settings.column_width+'px',"margin-right":'0px'});			
			}
		});
		
		$('#GridLayout ul li ul li').each(function(i)
		{
			if (i % gridLayoutObj.grid_settings.subcolumn_count < (gridLayoutObj.grid_settings.subcolumn_count-1))
			{
				$(this).css({width:gridLayoutObj.grid_settings.subcolumn_width+'px',"margin-right":gridLayoutObj.grid_settings.column_gutter+'px'});
			}
			else
			{
				$(this).css({width:gridLayoutObj.grid_settings.subcolumn_width+'px',"margin-right":'0px'});			
			}
		});
	}
};

GridLayout.prototype.calculateSubColumnWidth = function ()
{
	this.grid_settings.subcolumn_width = (this.grid_settings.column_width/this.grid_settings.subcolumn_count) - (((this.grid_settings.subcolumn_count-1)*this.grid_settings.column_gutter)/this.grid_settings.subcolumn_count);
};

GridLayout.prototype.calculateTotalWidth = function ()
{
	this.grid_settings.total_width = (this.grid_settings.column_width*this.grid_settings.column_count) + ((this.grid_settings.column_count-1)*this.grid_settings.column_gutter);
};

GridLayout.prototype.gridHeight = function ()
{
	var arrWinSizeAndScroll = this.getWinSizeAndScroll();

	$('#GridLayout ul li').css({height:(arrWinSizeAndScroll[1] + arrWinSizeAndScroll[3] + "px")});
	
	if (this.grid_settings.align == 'center')
	{
		// #JBD
		if(this.grid_settings.percent == true) { // Test for percentage setting
			// We're using percentages
			if(this.grid_settings.min_width != null && document.body.offsetWidth*(this.grid_settings.width/100) < this.grid_settings.min_width )
			{ // Test for a min-width
				// We're using a min-width.
				$('#GridLayout').css({left: (document.body.offsetWidth-this.grid_settings.min_width)/2+"px"});
			}
			else
			{
				// We're not using a min-width.
				$('#GridLayout').css({left: ((100-this.grid_settings.width)/2)+"%"});
			}
		}
		else
		{
			// We're not using percentages, default to pixel widths
			$('#GridLayout').css({left:(arrWinSizeAndScroll[0] + "px")});
		}
	}
	else if (this.grid_settings.align == 'right')
	{
		$('#GridLayout').css({right:"0px",left:"auto"});	    
	}
	else
	{
		$('#GridLayout').css({left:"0px",right:"auto"});	    	    
	}
};

GridLayout.prototype.getWinSizeAndScroll = function ()
{
	var intWidth   = document.body.offsetWidth;
	var intHeight  = (typeof window.innerHeight != "undefined")? window.innerHeight : (document.documentElement && document.documentElement.clientHeight > 0)? document.documentElement.clientHeight : document.body.clientHeight;		
	var intXScroll = (typeof window.pageXOffset != "undefined")? window.pageXOffset : document.body.scrollLeft;		
	var intYScroll = (typeof window.window.pageYOffset != "undefined")? window.window.pageYOffset : (document.documentElement && document.documentElement.scrollTop > 0)? document.documentElement.scrollTop : document.body.scrollTop;

	var offsetLeft = intWidth/2 - this.grid_settings.total_width/2;

	return [offsetLeft, intHeight, intXScroll, intYScroll];
};

GridLayout.prototype.checkKeyPress = function (e)
{
	var keynum;
	var keychar;
	var numcheck;
        
	if(e.keyCode) // IE
	{
		keynum = e.keyCode;
	}
	else if(e.which) // Netscape/Firefox/Opera
	{
		keynum = e.which;
	}
	
    // alert(keynum);
		
	keychar  = String.fromCharCode(keynum);
	
	if (keynum == 103 || keynum == 71 || keynum == 7)
		$('#GridLayout').toggle();
}
