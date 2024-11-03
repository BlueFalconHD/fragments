-- breadcrumb.frag

builder {
	breadcrumb = function()
		-- 'this' is an instance of a special fragment metaclass. This class includes functions		
		local heirarchy = this:getPageHeirarchy()
		local breadcrumb = ""
		
		for _, item in ipairs(hierarchy) do
		    local pagePath = item:getPagePath()
		    local pageName = item.meta:get("pageName", "NAMELESS PAGE")
		
		    breadcrumb = breadcrumb .. string.format('<a href="%s">%s</a> / ', pagePath, pageName)
		end
		
		return breadcrumb:sub(1, -4)
	end,
}

===

*{breadcrumb}

Welcome to ${siteName}.
Today's date is *{date}.

@{footer}
